package main

import (
	"net/http"
        "base"
	"crypto/tls"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"os"
	"log"
	"strings"
	"io/ioutil"
	"path"
	"html/template"
	"bytes"
	"fmt"
	"net/url"
	"net/http/httputil"
	"net"
	"time"
	"sync"
	"encoding/json"
	"golang.org/x/crypto/acme/autocert"
)


var tlsCertPath = os.Getenv("TLS_CERT_PATH")
var tlsKeyPath = os.Getenv("TLS_KEY_PATH")
var DNSDomain = os.Getenv("DNS_DOMAIN")
var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
var TTYDHostConsole = os.Getenv("TTYD_HOST_CONSOLE_PORT")
var TTYDem100Bios = os.Getenv("TTYD_EM100_BIOS_PORT")
var TTYDem100BMC = os.Getenv("TTYD_EM100_BMC_PORT")
var TTYDOSLoader = os.Getenv("TTYD_OS_LOADER")
var CTRLIp = os.Getenv("CTRL_IP")
var CTRLTcpPort = os.Getenv("CTRL_TCPPORT")
var certStorage = os.Getenv("CERT_STORAGE")
var ExpectedBMCIp = os.Getenv("EXPECT_BMC_IP")
var credentialUri = os.Getenv("CREDENTIALS_URI")
var credentialPort = os.Getenv("CREDENTIALS_TCPPORT")
var compileUri = os.Getenv("COMPILE_URI")
var compileTcpPort = os.Getenv("COMPILE_TCPPORT")
var StorageURI = os.Getenv("STORAGE_URI")
var StorageTCPPORT = os.Getenv("STORAGE_TCPPORT")

type serverEntry struct {
        servername string
        ip string
        tcpPort string
	compileIp string
        bmcIp string
	currentOwner string
	queue int
	expiration time.Time
}

type serversList struct {
	servers []serverEntry
	mux sync.Mutex
}

var ciServers serversList

// httpsRedirect redirects http requests to https
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
    http.Redirect(
        w, r,
        "https://"+r.Host+r.URL.String(),
        http.StatusMovedPermanently,
    )
}

func ShiftPath(p string) (head, tail string) {
    p = path.Clean("/" + p)
    i := strings.Index(p[1:], "/") + 1
    if i <= 0 {
        return p[1:], "/"
    }
    return p[1:i], p[i:]
}

func checkAccess(w http.ResponseWriter, r *http.Request, login string, command string) (bool){
	switch command {
		case "getToken":
				if ( r.Method == http.MethodGet || r.Method == http.MethodPost ) {
					return true
				} else {
					return false
				}
		case "validateUser":
				return true
		case "resetPassword":
				return true
		case "generatePasswordLnkRst":
				return true
		case "createUser":
				return true
	}
        if ( r.Header.Get("Authorization") != "" ) {
		var method string
		switch r.Method {
			case http.MethodGet:
				method = "GET"
			case http.MethodPut:
				method = "PUT"
			case http.MethodPost:
				method = "POST"
			case http.MethodDelete:
				method = "DELETE"
		}
                // Is this an AWS request ?
                words := strings.Fields(r.Header.Get("Authorization"))
                if ( words[0] == "OSF" ) {
                        // Let's dump the various content
                        keys := strings.Split(words[1],":")
                        // We must retrieve the secret key used for encryption and calculate the header
                        // if everything is ok (aka our computed value match) we are good

		        username := login

			result:=base.HTTPGetRequest("http://"+r.Host+":9100"+"/user/"+username+"/userGetInternalInfo")

			var return_data *base.User
			return_data = new(base.User)
                        json.Unmarshal([]byte(result),return_data)

			// I am getting the Secret Key and the Nickname
                        stringToSign := method + "\n\n"+r.Header.Get("Content-Type")+"\n"+r.Header.Get("myDate")+"\n"+r.URL.Path

			secretKey := return_data.TokenSecret
			nickname := username
			if ( nickname != login ) {
				return false
			}
                        mac := hmac.New(sha1.New, []byte(secretKey))
                        mac.Write([]byte(stringToSign))
                        expectedMAC := mac.Sum(nil)
                        if ( base64.StdEncoding.EncodeToString(expectedMAC) == keys[1] ) {
				return true
                        }
                }
	}
	return false
}

func user(w http.ResponseWriter, r *http.Request) {
	
        var command string
        entries := strings.Split(strings.TrimSpace(r.URL.Path[1:]), "/")
        var login string

        // The login is always accessible
        if ( len(entries) > 2 ) {
                command = entries[2]
                login = entries[1]
        }

	if ( !checkAccess(w, r, login, command)  ) {
		w.Write([]byte("Access denied"))
		return
	}

	// parse the url
	url, _ := url.Parse("http://"+credentialUri+credentialPort)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	r.URL.Host = "http://"+r.Host+":9100"

	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w , r)
}

func home(w http.ResponseWriter, r *http.Request) {

	// The cookie allow us to track the current
	// user on the node
        cookie, cookieErr := r.Cookie("osfci_cookie")
	cacheIndex := -1
	// We have to find the entry into the cache
	// if the cookie exist and return a Value

	if ( cookieErr == nil ) {
		if ( cookie.Value != "" ) {
		       for i, _ := range ciServers.servers {
	                        if ( ciServers.servers[i].currentOwner == cookie.Value ) {
					// Before indexing we must validate that the server is still ours 
					if ( time.Now().After(ciServers.servers[i].expiration) ) {
						ciServers.mux.Lock()
						ciServers.servers[i].expiration = time.Now();
       		                                ciServers.servers[i].currentOwner = ""
						// We have to reset the associated compile node and associated ctrl node
						client := &http.Client{}
                                                var req *http.Request
                                                req, _ = http.NewRequest("GET","http://"+ciServers.servers[i].ip+ciServers.servers[i].tcpPort+"/poweroff", nil)
                                                _, _  = client.Do(req)
						client = &http.Client{}
		                                req, _ = http.NewRequest("GET","http://"+ciServers.servers[i].compileIp+"/cleanUp", nil)
               			                 _, _  = client.Do(req)
						ciServers.mux.Unlock()
					} else {
						cacheIndex=i
					}
                                }
                       }
		}
	}

	head, tail := ShiftPath( r.URL.Path)
	if ( head == "ci" ) {
		head,_ = ShiftPath(tail)
	}
	// Some commands are superseed by a username so we shall identify
	// if that is the case. If the command is unknown then we can assume
	// we are getting a username as a head parameter and must get the 
	// remaining part

	// If the request is different than a getServer
	// We must be sure that the end user still has an active server
	// If that is not the case we deny the request
	// And need to re route the end user to an end of session
	switch ( head ) {
		case "getServer":
			// We need to have a valid cookie and associated Public Key / Private Key otherwise
			// We can't request a server
			if ( cookieErr == nil ) {
				if ( cookie.Value != "" ) {
					// To do so I must sent the cookie value to the user API and
					// get a respond. If it is gone we must denied the demand
					 type returnValue struct {
       	                                 Servername string
       	                                 Waittime string
						Queue string
						RemainingTime string
       		                         }
       		                         var myoutput returnValue	
					actualTime := time.Now().Add(time.Second*3600*365*10)
					index := 0
					ciServers.mux.Lock()
					for i, _ := range ciServers.servers { 
						if ( time.Now().After(ciServers.servers[i].expiration) ) {
							// the server is available we can allocate it
							ciServers.servers[i].expiration = time.Now().Add(time.Second*time.Duration(base.MaxServerAge))
							ciServers.servers[i].currentOwner = cookie.Value
							ciServers.mux.Unlock()
	
							myoutput.Servername = ciServers.servers[i].servername
							myoutput.Waittime = "0"
							myoutput.RemainingTime = fmt.Sprintf("%d",base.MaxServerAge)
							return_data,_ := json.Marshal(myoutput)
							if ( ciServers.servers[i].queue > 0 ) {
								ciServers.servers[i].queue = ciServers.servers[i].queue - 1
							}
							// We probably need to turn it off just to clean it
							client := &http.Client{}
			                                var req *http.Request
			                                req, _ = http.NewRequest("GET","http://"+ciServers.servers[i].ip+ciServers.servers[i].tcpPort+"/poweroff", nil)
			                                _, _  = client.Do(req)
							client = &http.Client{}
                                                        req, _ = http.NewRequest("GET","http://"+ciServers.servers[i].compileIp+"/cleanUp", nil)
                                                        _, _  = client.Do(req)
							w.Write([]byte(return_data))
							return
						} else {
							// We can check also if the user is just coming back ?
							// their could be a case where the user reloaded it's session
							// we can bring it back the server for his own usage
							if ( ciServers.servers[i].currentOwner == cookie.Value ) {
								// let's give it back to the user after a cleaning
								client := &http.Client{}
				                                var req *http.Request
                           					req, _ = http.NewRequest("GET","http://"+ciServers.servers[i].ip+ciServers.servers[i].tcpPort+"/poweroff", nil)
				                                _, _  = client.Do(req)
								client = &http.Client{}
       		                                                req, _ = http.NewRequest("GET","http://"+ciServers.servers[i].compileIp+"/cleanUp", nil)
                                            	            	_, _  = client.Do(req)
								myoutput.Servername = ciServers.servers[i].servername
	                                                        myoutput.Waittime = "0"
								myoutput.RemainingTime = fmt.Sprintf("%d", ciServers.servers[i].expiration.Unix() - time.Now().Unix())
	                                                        return_data,_ := json.Marshal(myoutput)
	                                                        if ( ciServers.servers[i].queue > 0 ) {
	                                                                ciServers.servers[i].queue = ciServers.servers[i].queue - 1
	                                                        }
								ciServers.mux.Unlock()
	                                                        w.Write([]byte(return_data))
	                                                        // We probably need to turn it off just to clean it
	                                                        return
							}
						}
						// used to calculate potential wait time
						if ( actualTime.After(ciServers.servers[i].expiration) ) {
							actualTime = ciServers.servers[i].expiration
							index = i
						}
						
					}
					myoutput.Servername = ""
					remainingTime := actualTime.Sub(time.Now())
					myoutput.Waittime = fmt.Sprintf("%.0f", remainingTime.Seconds())
					myoutput.Queue = fmt.Sprintf("%d",ciServers.servers[index].queue)
					ciServers.servers[index].queue = ciServers.servers[index].queue + 1	
					ciServers.mux.Unlock()
					myoutput.RemainingTime = fmt.Sprintf("%d",0)
					return_data,_ := json.Marshal(myoutput)
					w.Write([]byte(return_data))
				}
			}
		case "stopServer":
			// We must get the server name
			_,tail := ShiftPath(tail)
			servername,_ := ShiftPath(tail)
			// Ok we must look for this server into the ciServer list
			// we must validate that the cookie if the right one
			if ( cookieErr == nil ) {
				ciServers.mux.Lock()
				for i, _ := range ciServers.servers {
					if ( ciServers.servers[i].servername == servername ) {
						if ( ciServers.servers[i].currentOwner == cookie.Value ) {
							// Ok we can free the server
							// This is done by resetting the expiration
							ciServers.servers[i].expiration = time.Now();
							ciServers.servers[i].currentOwner = ""
							client := &http.Client{}
	                                                var req *http.Request
       		                                        req, _ = http.NewRequest("GET","http://"+ciServers.servers[i].compileIp+compileTcpPort+"/cleanUp", nil)
                	                                _, _  = client.Do(req)
						}
					}
				}
				ciServers.mux.Unlock()
			}
		case "getosinstallers":
			// Must get a directory content from the storage backend if there is no further option
			// if their is an option (aka a file name), it means that we have to inform the
			// current user controller to load that file
			// and startup the associated ttyd
			path := strings.Split( r.URL.Path, "/" )
			if ( len(path) < 3 ) {
		                http.Error(w, "401 Malformed URI", 401)
        		}
		        if ( len(path) == 4 && path[3] == "" ) {
				// So we forward the request to the storage backend
				client := &http.Client{}
				var req *http.Request
	                        req, _ = http.NewRequest("GET","http://"+StorageURI+StorageTCPPORT+"/distros/", nil)
	                        resp, _  := client.Do(req)
				defer resp.Body.Close()
				body,_ := ioutil.ReadAll(resp.Body)
				w.Write([]byte(body));
			} else {
				// we got a file name we have to forward the request to the controller node
				// we must  request to the relevant test server
                                client := &http.Client{}
                                var req *http.Request

                                req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/getosinstallers/"+path[3], nil)
                                _, _  = client.Do(req)
			}
		case "bmcup":
		       bmcIp := ""
			var Up string
			if ( cacheIndex != -1 ) {
				bmcIp = ciServers.servers[cacheIndex].bmcIp
                      	}
			if ( bmcIp != "" ) {
			        conn, err := net.DialTimeout("tcp", bmcIp+":443", 220*time.Millisecond)
			        if ( err == nil ) {
			                conn.Close()
					// The controller is up				
					Up = "1"
			        } else {
					Up = "0"
				}
			} else {
				Up = "0"
			}
			return_value,_ := json.Marshal(Up)
			w.Write([]byte(return_value))
		case "console":
			if ( cacheIndex != -1 ) {
				fmt.Printf("Console request\n");
			        url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDHostConsole)
			        proxy := httputil.NewSingleHostReverseProxy(url)
			        r.URL.Host = "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDHostConsole
				filePath :=  strings.Split(tail,"/")
				r.URL.Path = "/"
				if ( len(filePath) > 2 ) {
					r.URL.Path = r.URL.Path + filePath[2]
				}
			        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
				proxy.ServeHTTP(w , r)
			}
		case "isEmulatorsPool":
			if ( cacheIndex != -1 ) {
				url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort)
	                        proxy := httputil.NewSingleHostReverseProxy(url)
                                r.URL.Path = "/isEmulatorsPool"
                                r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                                proxy.ServeHTTP(w , r)
			}
		case "resetEmulator":				
			if ( cacheIndex != -1 ) {
                                url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort)
                                proxy := httputil.NewSingleHostReverseProxy(url)
                                r.URL.Path = "/resetEmulator" + tail
				fmt.Printf(r.URL.Path)
                                r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                                proxy.ServeHTTP(w , r)
                        }
		case "smbiosconsole":
			if ( cacheIndex != -1 ) {
	                        url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDem100Bios)
	                        proxy := httputil.NewSingleHostReverseProxy(url)
	                        r.URL.Host = "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDem100Bios
	                        filePath :=  strings.Split(tail,"/")
	                        r.URL.Path = "/"
	                        if ( len(filePath) > 2 ) {
	                                r.URL.Path = r.URL.Path + filePath[2]
	                        }
	                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	                        proxy.ServeHTTP(w , r)
			}
		case "smbiosbuildconsole":
			if ( cacheIndex != -1 ) {
	                        url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].compileIp+":7681")
	                        proxy := httputil.NewSingleHostReverseProxy(url)
	                        r.URL.Host = "http://"+ciServers.servers[cacheIndex].compileIp+TTYDem100Bios
	                        filePath :=  strings.Split(tail,"/")
	                        r.URL.Path = "/"
	                        if ( len(filePath) > 2 ) {
	                                r.URL.Path = r.URL.Path + filePath[2]
	                        }
	                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	                        proxy.ServeHTTP(w , r)
			}
                case "bmcbuildconsole":
                        if ( cacheIndex != -1 ) {
                                url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].compileIp+":7682")
                                proxy := httputil.NewSingleHostReverseProxy(url)
                                r.URL.Host = "http://"+ciServers.servers[cacheIndex].compileIp+TTYDem100BMC
                                filePath :=  strings.Split(tail,"/")
                                r.URL.Path = "/"
                                if ( len(filePath) > 2 ) {
                                        r.URL.Path = r.URL.Path + filePath[2]
                                }
                                r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                                proxy.ServeHTTP(w , r)
                        }
		case "osloaderconsole":
                        if ( cacheIndex != -1 ) {
                                url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDOSLoader)
                                proxy := httputil.NewSingleHostReverseProxy(url)
                                r.URL.Host = "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDOSLoader
                                filePath :=  strings.Split(tail,"/")
                                r.URL.Path = "/"
                                if ( len(filePath) > 2 ) {
                                        r.URL.Path = r.URL.Path + filePath[2]
                                }
                                r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                                proxy.ServeHTTP(w , r)
                        }
		case "poweron":
			if ( cacheIndex != -1 ) {
				fmt.Printf("Poweron request\n");
				client := &http.Client{}
	                        var req *http.Request
	                        req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/poweron", nil)
	                        _, _  = client.Do(req)
			}
		case "poweroff":
			if ( cacheIndex != -1 ) {
				fmt.Printf("Poweroff request\n");
				client := &http.Client{}
	                        var req *http.Request
	                        req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/poweroff", nil)
	                        _, _  = client.Do(req)
			}
		case "bmcconsole":
			if ( cacheIndex != -1 ) {
	                        url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDem100BMC)
	                        proxy := httputil.NewSingleHostReverseProxy(url)
	                        r.URL.Host = "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+TTYDem100BMC
	                        filePath :=  strings.Split(tail,"/")
	                        r.URL.Path = "/"
	                        if ( len(filePath) > 2 ) {
	                                r.URL.Path = r.URL.Path + filePath[2]
	                        }
	                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	                        proxy.ServeHTTP(w , r)
			}
		case "startbmc":
			if ( cacheIndex != -1 ) {
				// we must forward the request to the relevant test server
				client := &http.Client{}
				var req *http.Request
				req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/startbmc", nil)
			        _, _  = client.Do(req)
				client = &http.Client{}
	                        req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/startbmcconsole", nil)
	                        _, _  = client.Do(req)
			}
		case "startsmbios":
			if ( cacheIndex != -1 ) {
				// we must forward the request to the relevant test server
	                        client := &http.Client{}
	                        var req *http.Request
	                        req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/startsmbios", nil)
	                        _, _  = client.Do(req)
			}
		case "js":
			b, _ := ioutil.ReadFile(staticAssetsDir+tail) // just pass the file name
                        w.Write(b)
		case "html":
			b, _ := ioutil.ReadFile(staticAssetsDir+tail) // just pass the file name
                        w.Write(b)
		case "css":
			b, _ := ioutil.ReadFile(staticAssetsDir+tail) // just pass the file name
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
                        w.Write(b)
		case "images":
			b, _ := ioutil.ReadFile(staticAssetsDir+tail) // just pass the file name
			w.Header().Set("Content-Type", "image/png")
			w.Write(b)
		case "mp4":
			b, _ := ioutil.ReadFile(staticAssetsDir+tail) // just pass the file name
                        w.Header().Set("Content-Type", "video/mp4")
                        w.Write(b)
		case "bmcfirmware":
			if ( cacheIndex != -1 ) {
				// We must forward the request
				fmt.Printf("Forward bmcfirmware upload\n");
	                        url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort)
	                        proxy := httputil.NewSingleHostReverseProxy(url)
	                        r.URL.Host = "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort
				_, tail = ShiftPath( r.URL.Path)
				path :=  strings.Split(tail,"/")
	                        r.URL.Path = "/bmcfirmware/"+path[2]
	                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	                        proxy.ServeHTTP(w , r)
			}
		case "biosfirmware":
			if ( cacheIndex != -1 ) {
				// We must forward the request
	                        fmt.Printf("Forward biosfirmware upload\n");
	                        url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort)
	                        proxy := httputil.NewSingleHostReverseProxy(url)
				r.URL.Host = "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort
				_, tail = ShiftPath( r.URL.Path)
                                path :=  strings.Split(tail,"/")
	                        r.URL.Path = "/biosfirmware/"+path[2]
	                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	                        proxy.ServeHTTP(w , r)
			}
		case "buildbiosfirmware":
			if ( cacheIndex != -1 ) {
				_, tail = ShiftPath( r.URL.Path)
	                        keys := strings.Split(tail,"/")
				login := keys[2]
				command := keys[1]
			        if ( !checkAccess(w, r, login, command)  ) {
       	 		        w.Write([]byte("Access denied"))
		               		 return
       		 		}
				// We have to forward the request to the compile server
				// which will start the compilation process and return
				// the code to connect to the ttyd daemon
				fmt.Printf("Forward biosfirmware build\n");
       		                url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].compileIp+compileTcpPort)
       		                proxy := httputil.NewSingleHostReverseProxy(url)
       		                r.URL.Host = "http://"+ciServers.servers[cacheIndex].compileIp+compileTcpPort
				fmt.Printf("Tail %s\n",tail)
       		                r.URL.Path = tail
       		                r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
       		                proxy.ServeHTTP(w , r)
			}
		case "buildbmcfirmware":
                        if ( cacheIndex != -1 ) {
                                _, tail = ShiftPath( r.URL.Path)
                                keys := strings.Split(tail,"/")
                                login := keys[2]
                                command := keys[1]
                                if ( !checkAccess(w, r, login, command)  ) {
                                w.Write([]byte("Access denied"))
                                         return
                                }
                                // We have to forward the request to the compile server
                                // which will start the compilation process and return
                                // the code to connect to the ttyd daemon
                                fmt.Printf("Forward bmcfirmware build\n");
                                url, _ := url.Parse("http://"+ciServers.servers[cacheIndex].compileIp+compileTcpPort)
                                proxy := httputil.NewSingleHostReverseProxy(url)
                                r.URL.Host = "http://"+ciServers.servers[cacheIndex].compileIp+compileTcpPort
                                fmt.Printf("Tail %s\n",tail)
                                r.URL.Path = tail
                                r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                                proxy.ServeHTTP(w , r)
                        }
		case "loadbuiltsmbios":
			if ( cacheIndex != -1 ) {
				// we must tell to the controller node that he needs to download the BIOS
				// from our user from the storage node and to start the em100
				_, tail = ShiftPath( r.URL.Path)
       		                keys := strings.Split(tail,"/")
       		                login := keys[2]
       		                client := &http.Client{}
       		                var req *http.Request
       		                req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/loadfromstoragesmbios/"+login, nil)
       		                _, _  = client.Do(req)
			}
		case "loadbuiltopenbmc":
			if ( cacheIndex != -1 ) {
                                // we must tell to the controller node that he needs to download the BIOS
                                // from our user from the compile node and to start the em100
                                _, tail = ShiftPath( r.URL.Path)
                                keys := strings.Split(tail,"/")
                                login := keys[2]
                                client := &http.Client{}
                                var req *http.Request
                                req, _ = http.NewRequest("GET","http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/loadfromstoragebmc/"+login, nil)
                                _, _  = client.Do(req)
                        }
		case "":
                        b, _ := ioutil.ReadFile(staticAssetsDir+"/html/homepage.html") // just pass the file name
                        // this is a potential template file we need to replace the http field
                        // by the calling r.Host
                        t := template.New("my template")
                        buf := &bytes.Buffer{}
                        t.Parse(string(b))
                        t.Execute(buf, r.Host+"/ci/")
                        fmt.Fprintf(w, buf.String())
		default:
	}
}

func bmcweb(w http.ResponseWriter, r *http.Request){
	// Let's print the session ID
        cookie, err := r.Cookie("osfci_cookie")


	// If the request is for a favicon.ico file we are just returning
	// we do not offer such icon currently ;)
	head, _ := ShiftPath( r.URL.Path)
	if ( head == "favicon.ico" ) {
		return
	}

	// We must validate if the user is coming with a cookie
	// if yes we must look to which server it is allocated
	// if it is not allocated to any then we probably need to reroute him
	// to the homepage !

	bmcIp := ""
	if ( err == nil ) {
		if ( cookie.Value != "" ) {
			// We must get the IP address from the cache
			for i, _ := range ciServers.servers {
				if ( ciServers.servers[i].currentOwner == cookie.Value ) {
					if ( time.Now().Before(ciServers.servers[i].expiration) ) {
						// We still own the server and we can go to the BMC
						bmcIp = ciServers.servers[i].bmcIp
					}
				}
			}
		} else {
			if ( DNSDomain != "" ) {
       	                 http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
       	         }
       	         return
		}
	} else {
		if ( DNSDomain != "" ) {
                        http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
                }
                return
	}	
	if ( bmcIp == "" ) {
                if ( DNSDomain != "" ) {
                        http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
                }
                return
        }
	// We must know if iLo is started or not ?
	// if not then we have to reroute to the actual homepage
	// We can make a request to the website or
	conn, err := net.DialTimeout("tcp", bmcIp+":443", 220*time.Millisecond)
	if ( err != nil ) {
		if ( DNSDomain != "" ) {
			http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
		}
		return
	} else {
		conn.Close()
	}
	// Must specify the iLo Web address
	url, _ := url.Parse("https://"+bmcIp+":443")
	proxy := httputil.NewSingleHostReverseProxy(url)
	var InsecureTransport http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		}).Dial,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			TLSHandshakeTimeout: 10 * time.Second,
	}
	// Our OpenBMC has a self signed certificate
	proxy.Transport = InsecureTransport
	// Internal gateway IP address
	// Must reroute on myself and port 443
        url, _ = url.Parse("http://"+r.Header.Get("Host"))
	r.URL.Host = "https://"+url.Hostname()+":443/"
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	proxy.ServeHTTP(w , r)

}

func main() {
    print("=============================== \n")
    print("| Starting frontend           |\n")
    print("| Development version -       |\n")
    print("| Private use only            |\n")
    print("=============================== \n")
    print(" Please do not forget to set TLS_CERT_PATH/TLS_KEY_PATH/STATIC_ASSETS_DIR to there relevant path\n")

    mux := http.NewServeMux()

    // Highest priority must be set to the signed request
    mux.HandleFunc("/ci/",home)
    mux.HandleFunc("/user/", user)
    mux.HandleFunc("/",bmcweb)

    // We must build our server pool for the moment
    // This is define by the environment variable
    // But this could be done by a registration mechanism later
    var newEntry serverEntry

    newEntry.servername = "dl360"
    newEntry.ip=CTRLIp
    newEntry.tcpPort=CTRLTcpPort
    newEntry.compileIp=compileUri
    newEntry.currentOwner=""
    // the server is expired
    newEntry.expiration = time.Now()
    // by the way its bmc interface is
    newEntry.bmcIp=ExpectedBMCIp 
    newEntry.queue = 0

    ciServers.mux.Lock()
    ciServers.servers = append(ciServers.servers, newEntry)
    ciServers.mux.Unlock()

    if ( DNSDomain != "" ) {
        // if DNS_DOMAIN is set then we run in a production environment
        // we must get the directory where the certificates will be stored
        certManager := autocert.Manager{
                Prompt: autocert.AcceptTOS,
                Cache:  autocert.DirCache(certStorage),
                HostPolicy: autocert.HostWhitelist(DNSDomain),
        }

        server := &http.Server{
                Addr:    ":443",
                Handler: mux,
                ReadTimeout:  600 * time.Second,
                WriteTimeout: 600 * time.Second,
                IdleTimeout:  120 * time.Second,
                TLSConfig: &tls.Config{
                        GetCertificate: certManager.GetCertificate,
                },
        }

        go func() {
        h := certManager.HTTPHandler(nil)
                log.Fatal(http.ListenAndServe(":http", h))
        }()

        server.ListenAndServeTLS("", "")
     } else {
    		go http.ListenAndServe(":80", http.HandlerFunc(httpsRedirect))
	    	// Launch TLS server
	    	log.Fatal(http.ListenAndServeTLS(":443", tlsCertPath, tlsKeyPath, mux))
     }
}
