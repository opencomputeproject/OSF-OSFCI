// OSFCI Server module

package main

import (
	"base/base"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

var tlsCertPath string
var tlsKeyPath string

//DNSDomain is read from config
var DNSDomain string
var staticAssetsDir string

//TTYDHostConsole is read from config
var TTYDHostConsole string

//TTYDem100Bios is read from config
var TTYDem100Bios string

//TTYDem100BMC is read from config
var TTYDem100BMC string

// TTYDOSLoader is read from config
var TTYDOSLoader string

var certStorage string

// ExpectedBMCIp is read from config
var credentialURI string
var credentialPort string
var compileURI string
var compileTCPPort string

//StorageURI is read from config
var StorageURI string

//StorageTCPPORT is read from config
var StorageTCPPORT string

//MaxServerAge is read from config
var MaxServerAge int

type serverProduct struct {
	Product string
	Brand   string
	Active  int
}

var ciServersProducts []serverProduct

type serverEntry struct {
	servername   string
	ip           string
	tcpPort      string
	compileIP    string
	bmcIP        string
	currentOwner string
	gitToken     string
	queue        int
	expiration   time.Time
	ProductIndex int
}

type serversList struct {
	servers []serverEntry
	mux     sync.Mutex
}

var ciServers serversList

//Initialize the config variables
func initServerconfig() error {
	viper.SetConfigName("gatewayconf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/production/config/")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	tlsCertPath = viper.GetString("TLS_CERT_PATH")
	tlsKeyPath = viper.GetString("TLS_KEY_PATH")

	//DNSDomain set from config file
	DNSDomain = viper.GetString("DNS_DOMAIN")
	staticAssetsDir = viper.GetString("STATIC_ASSETS_DIR")

	//TTYDHostConsole set from config file
	TTYDHostConsole = viper.GetString("TTYD_HOST_CONSOLE_PORT")

	//TTYDem100Bios set from config file
	TTYDem100Bios = viper.GetString("TTYD_EM100_BIOS_PORT")

	//TTYDem100BMC set from config file
	TTYDem100BMC = viper.GetString("TTYD_EM100_BMC_PORT")

	//TTYDOSLoader set from config file
	TTYDOSLoader = viper.GetString("TTYD_OS_LOADER")

	certStorage = viper.GetString("CERT_STORAGE")

	credentialURI = viper.GetString("CREDENTIALS_URI")
	credentialPort = viper.GetString("CREDENTIALS_TCPPORT")
	compileURI = viper.GetString("COMPILE_URI")
	compileTCPPort = viper.GetString("COMPILE_TCPPORT")

	//StorageURI set from config file
	StorageURI = viper.GetString("STORAGE_URI")

	//StorageTCPPORT set from config file
	StorageTCPPORT = viper.GetString("STORAGE_TCPPORT")

	//MaxServerAge set from config file: currently 3600 -> 60 minutes
	MaxServerAge = viper.GetInt("MAX_SERVER_AGE")
	return nil
}

// httpsRedirect redirects http requests to https
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
		w, r,
		"https://"+r.Host+r.URL.String(),
		http.StatusMovedPermanently,
	)
}

// ShiftPath cleans up path
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func checkAccess(w http.ResponseWriter, r *http.Request, login string, command string) bool {
	switch command {
	case "get_token":
		return r.Method == http.MethodGet || r.Method == http.MethodPost
	case "validate_user":
		return true
	case "reset_password":
		return true
	case "generate_password_lnk_rst":
		return true
	case "create_user":
		return true
	}
	if r.Header.Get("Authorization") != "" {
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
		if words[0] == "OSF" {
			// Let's dump the various content
			keys := strings.Split(words[1], ":")
			// We must retrieve the secret key used for encryption and calculate the header
			// if everything is ok (aka our computed value match) we are good

			username := login

			result := base.HTTPGetRequest("http://" + r.Host + ":9100" + "/user/" + username + "/userGetInternalInfo")

			var returnData *base.User
			returnData = new(base.User)
			json.Unmarshal([]byte(result), returnData)

			// I am getting the Secret Key and the Nickname
			stringToSign := method + "\n\n" + r.Header.Get("Content-Type") + "\n" + r.Header.Get("myDate") + "\n" + r.URL.Path

			secretKey := returnData.TokenSecret
			nickname := username
			if nickname != login {
				return false
			}
			mac := hmac.New(sha1.New, []byte(secretKey))
			mac.Write([]byte(stringToSign))
			expectedMAC := mac.Sum(nil)
			if base64.StdEncoding.EncodeToString(expectedMAC) == keys[1] {
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
	base.Zlog.Infof("Length of entries: %d", len(entries))
	login = entries[1]
	if len(entries) > 2 {
		command = entries[2]
	}

	if !checkAccess(w, r, login, command) {
		w.Write([]byte("Access denied"))
		base.Zlog.Infof("Access denied: %s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		return
	}

	// parse the url
	url, _ := url.Parse("http://" + credentialURI + credentialPort)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	r.URL.Host = "http://" + r.Host + ":9100"

	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

	base.Zlog.Infof("Access: %s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, r)
}

func home(w http.ResponseWriter, r *http.Request) {
	// The cookie allow us to track the current
	// user on the node
	cookie, cookieErr := r.Cookie("osfci_cookie")
	cacheIndex := -1
	// We have to find the entry into the cache
	// if the cookie exist and return a Value

	if cookieErr == nil {
		if cookie.Value != "" {
			for i := range ciServers.servers {
				if ciServers.servers[i].currentOwner == cookie.Value {
					// Before indexing we must validate that the server is still ours
					if time.Now().After(ciServers.servers[i].expiration) {
						ciServers.mux.Lock()
						ciServers.servers[i].expiration = time.Now()
						ciServers.servers[i].currentOwner = ""
						ciServers.servers[i].gitToken = ""
						// We have to reset the associated compile node and associated ctrl node
						client := &http.Client{}
						var req *http.Request
						req, _ = http.NewRequest("GET", "http://"+ciServers.servers[i].ip+ciServers.servers[i].tcpPort+"/power_off", nil)
						_, _ = client.Do(req)
						client = &http.Client{}
						req, _ = http.NewRequest("GET", "http://"+ciServers.servers[i].compileIP+"/clean_up", nil)
						_, _ = client.Do(req)
						ciServers.mux.Unlock()
					} else {
						cacheIndex = i
					}
				}
			}
		}
	}

	head, tail := ShiftPath(r.URL.Path)
	if head == "ci" {
		head, _ = ShiftPath(tail)
	}
	// Some commands are superseed by a username so we shall identify
	// if that is the case. If the command is unknown then we can assume
	// we are getting a username as a head parameter and must get the
	// remaining part

	// If the request is different than a get_server
	// We must be sure that the end user still has an active server
	// If that is not the case we deny the request
	// And need to re route the end user to an end of session
	switch head {
	case "get_server_models":
		var activeProducts []serverProduct
		for i := range ciServersProducts {
			if ciServersProducts[i].Active != 0 {
				activeProducts = append(activeProducts, ciServersProducts[i])
			}
		}
		returnData, err := json.Marshal(activeProducts)
		if err != nil {
			base.Zlog.Fatalf("JSON encoding error: %s", err.Error())
		}
		w.Write([]byte(returnData))
	case "get_server":
		base.Zlog.Infof("get_server: %s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		var serverTypeIndex int
		serverTypeIndex = -1
		_, tail := ShiftPath(tail)
		serverType, _ := ShiftPath(tail)
		for i := range ciServersProducts {
			if ciServersProducts[i].Product == serverType {
				serverTypeIndex = i
			}
		}
		// We need to have a valid cookie and associated Public Key / Private Key otherwise
		// We can't request a server
		if cookieErr == nil {
			if cookie.Value != "" {
				// To do so I must sent the cookie value to the user API and
				// get a respond. If it is gone we must denied the demand
				type returnValue struct {
					Servername    string
					Waittime      string
					Queue         string
					RemainingTime string
				}
				var myoutput returnValue
				actualTime := time.Now().Add(time.Second * 3600 * 365 * 10)
				index := 0
				ciServers.mux.Lock()
				for i := range ciServers.servers {
					if time.Now().After(ciServers.servers[i].expiration) {
						if ciServers.servers[i].ProductIndex == serverTypeIndex {
							// the server is available we can allocate it
							ciServers.servers[i].expiration = time.Now().Add(time.Second * time.Duration(MaxServerAge))
							ciServers.servers[i].currentOwner = cookie.Value
							ciServers.mux.Unlock()

							myoutput.Servername = ciServers.servers[i].servername
							myoutput.Waittime = "0"
							myoutput.RemainingTime = fmt.Sprintf("%d", MaxServerAge)
							returnData, _ := json.Marshal(myoutput)
							if ciServers.servers[i].queue > 0 {
								ciServers.servers[i].queue = ciServers.servers[i].queue - 1
							}
							// We probably need to turn it off just to clean it
							client := &http.Client{}
							var req *http.Request
							req, _ = http.NewRequest("GET", "http://"+ciServers.servers[i].ip+ciServers.servers[i].tcpPort+"/power_off", nil)
							_, _ = client.Do(req)
							client = &http.Client{}
							req, _ = http.NewRequest("GET", "http://"+ciServers.servers[i].compileIP+"/clean_up", nil)
							_, _ = client.Do(req)
							w.Write([]byte(returnData))
							return
						}
						// We can check also if the user is just coming back ?
						// their could be a case where the user reloaded it's session
						// we can bring it back the server for his own usage
						if ciServers.servers[i].currentOwner == cookie.Value {
							// let's give it back to the user after a cleaning
							client := &http.Client{}
							var req *http.Request
							req, _ = http.NewRequest("GET", "http://"+ciServers.servers[i].ip+ciServers.servers[i].tcpPort+"/power_off", nil)
							_, _ = client.Do(req)
							client = &http.Client{}
							req, _ = http.NewRequest("GET", "http://"+ciServers.servers[i].compileIP+"/clean_up", nil)
							_, _ = client.Do(req)
							myoutput.Servername = ciServers.servers[i].servername
							myoutput.Waittime = "0"
							myoutput.RemainingTime = fmt.Sprintf("%d", ciServers.servers[i].expiration.Unix()-time.Now().Unix())
							returnData, _ := json.Marshal(myoutput)
							if ciServers.servers[i].queue > 0 {
								ciServers.servers[i].queue = ciServers.servers[i].queue - 1
							}
							ciServers.mux.Unlock()
							w.Write([]byte(returnData))
							// We probably need to turn it off just to clean it
							return
						}
					}
					// used to calculate potential wait time
					if ciServers.servers[i].ProductIndex == serverTypeIndex {
						if actualTime.After(ciServers.servers[i].expiration) {
							actualTime = ciServers.servers[i].expiration
							index = i
						}
					}

				}
				myoutput.Servername = ""
				remainingTime := actualTime.Sub(time.Now())
				myoutput.Waittime = fmt.Sprintf("%.0f", remainingTime.Seconds())
				myoutput.Queue = fmt.Sprintf("%d", ciServers.servers[index].queue)
				ciServers.servers[index].queue = ciServers.servers[index].queue + 1
				ciServers.mux.Unlock()
				myoutput.RemainingTime = fmt.Sprintf("%d", 0)
				returnData, _ := json.Marshal(myoutput)
				w.Write([]byte(returnData))
			}
		}
	case "stop_server":
		// We must get the server name
		_, tail := ShiftPath(tail)
		servername, _ := ShiftPath(tail)
		// Ok we must look for this server into the ciServer list
		// we must validate that the cookie if the right one
		if cookieErr == nil {
			ciServers.mux.Lock()
			for i := range ciServers.servers {
				if ciServers.servers[i].servername == servername {
					if ciServers.servers[i].currentOwner == cookie.Value {
						// Ok we can free the server
						// This is done by resetting the expiration
						ciServers.servers[i].expiration = time.Now()
						ciServers.servers[i].currentOwner = ""
						ciServers.servers[i].gitToken = ""
						client := &http.Client{}
						var req *http.Request
						req, _ = http.NewRequest("GET", "http://"+ciServers.servers[i].compileIP+compileTCPPort+"/clean_up", nil)
						_, _ = client.Do(req)
					}
				}
			}
			ciServers.mux.Unlock()
		}
	case "get_os_installers":
		// Must get a directory content from the storage backend if there is no further option
		// if their is an option (aka a file name), it means that we have to inform the
		// current user controller to load that file
		// and startup the associated ttyd
		path := strings.Split(r.URL.Path, "/")
		if len(path) < 3 {
			http.Error(w, "401 Malformed URI", 401)
		}
		if len(path) == 4 && path[3] == "" {
			// So we forward the request to the storage backend
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET", "http://"+StorageURI+StorageTCPPORT+"/distros/", nil)
			resp, _ := client.Do(req)
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			w.Write([]byte(body))
		} else {
			// we got a file name we have to forward the request to the controller node
			// we must  request to the relevant test server
			client := &http.Client{}
			var req *http.Request

			req, _ = http.NewRequest("GET", "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/get_os_installers/"+path[3], nil)
			_, _ = client.Do(req)
		}
	case "bmc_up":
		bmcIP := ""
		var Up string
		if cacheIndex != -1 {
			bmcIP = ciServers.servers[cacheIndex].bmcIP
		}
		if bmcIP != "" {
			conn, err := net.DialTimeout("tcp", bmcIP+":443", 220*time.Millisecond)
			if err == nil {
				conn.Close()
				// The controller is up
				Up = "1"
			} else {
				Up = "0"
			}
		} else {
			Up = "0"
		}
		returnValue, _ := json.Marshal(Up)
		w.Write([]byte(returnValue))
	case "console":
		if cacheIndex != -1 {
			//fmt.Printf("Console request\n")
			base.Zlog.Infof("Console request")
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDHostConsole)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDHostConsole
			filePath := strings.Split(tail, "/")
			r.URL.Path = "/"
			if len(filePath) > 2 {
				r.URL.Path = r.URL.Path + filePath[2]
			}
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "is_running":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].compileIP + compileTCPPort)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].compileIP + compileTCPPort
			//fmt.Printf("Tail %s\n", tail)
			base.Zlog.Infof("Tail %s", tail)
			r.URL.Path = tail
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "is_emulators_pool":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Path = "/is_emulators_pool"
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "reset_emulator":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Path = tail
			//fmt.Printf(r.URL.Path)
			base.Zlog.Infof("%s", r.URL.Path)
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "smbios_console":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDem100Bios)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDem100Bios
			filePath := strings.Split(tail, "/")
			r.URL.Path = "/"
			if len(filePath) > 2 {
				r.URL.Path = r.URL.Path + filePath[2]
			}
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "smbios_build_console":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].compileIP + ":7681")
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].compileIP + TTYDem100Bios
			filePath := strings.Split(tail, "/")
			r.URL.Path = "/"
			if len(filePath) > 2 {
				r.URL.Path = r.URL.Path + filePath[2]
			}
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "bmc_build_console":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].compileIP + ":7682")
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].compileIP + TTYDem100BMC
			filePath := strings.Split(tail, "/")
			r.URL.Path = "/"
			if len(filePath) > 2 {
				r.URL.Path = r.URL.Path + filePath[2]
			}
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "os_loader_console":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDOSLoader)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDOSLoader
			filePath := strings.Split(tail, "/")
			r.URL.Path = "/"
			if len(filePath) > 2 {
				r.URL.Path = r.URL.Path + filePath[2]
			}
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "power_on":
		if cacheIndex != -1 {
			//fmt.Printf("Poweron request\n")
			base.Zlog.Infof("Poweron request")
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET", "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/power_on", nil)
			_, _ = client.Do(req)
		}
	case "power_off":
		if cacheIndex != -1 {
			//fmt.Printf("Poweroff request\n")
			base.Zlog.Infof("Poweroff request")
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET", "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/power_off", nil)
			_, _ = client.Do(req)
		}
	case "bmc_console":
		if cacheIndex != -1 {
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDem100BMC)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort + TTYDem100BMC
			filePath := strings.Split(tail, "/")
			r.URL.Path = "/"
			if len(filePath) > 2 {
				r.URL.Path = r.URL.Path + filePath[2]
			}
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "start_bmc":
		if cacheIndex != -1 {
			// we must forward the request to the relevant test server
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET", "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/start_bmc", nil)
			_, _ = client.Do(req)
		}
	case "start_smbios":
		if cacheIndex != -1 {
			// we must forward the request to the relevant test server
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET", "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/start_smbios", nil)
			_, _ = client.Do(req)
		}
	case "js":
		b, _ := ioutil.ReadFile(staticAssetsDir + tail) // just pass the file name
		w.Write(b)
	case "html":
		b, _ := ioutil.ReadFile(staticAssetsDir + tail) // just pass the file name
		w.Write(b)
	case "css":
		b, _ := ioutil.ReadFile(staticAssetsDir + tail) // just pass the file name
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(b)
	case "images":
		b, _ := ioutil.ReadFile(staticAssetsDir + tail) // just pass the file name
		w.Header().Set("Content-Type", "image/png")
		w.Write(b)
	case "mp4":
		b, _ := ioutil.ReadFile(staticAssetsDir + tail) // just pass the file name
		w.Header().Set("Content-Type", "video/mp4")
		w.Write(b)
	case "bmc_firmware":
		if cacheIndex != -1 {
			// We must forward the request
			//fmt.Printf("Forward bmcfirmware upload\n")
			base.Zlog.Infof("Forward bmc firmware upload")
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort
			_, tail = ShiftPath(r.URL.Path)
			path := strings.Split(tail, "/")
			r.URL.Path = "/bmc_firmware/" + path[2]
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "bios_firmware":
		if cacheIndex != -1 {
			// We must forward the request
			//fmt.Printf("Forward biosfirmware upload\n")
			base.Zlog.Infof("Forward bios firmware upload")
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].ip + ciServers.servers[cacheIndex].tcpPort
			_, tail = ShiftPath(r.URL.Path)
			path := strings.Split(tail, "/")
			r.URL.Path = "/bios_firmware/" + path[2]
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "git_token":
		if cacheIndex != -1 {
			_, tail = ShiftPath(r.URL.Path)
			keys := strings.Split(tail, "/")
			login := keys[2]
			command := keys[1]
			if !checkAccess(w, r, login, command) {
				w.Write([]byte("Access denied"))
				return
			}
			data := base.HTTPGetBody(r)
			ciServers.servers[cacheIndex].gitToken = string(data)
			//fmt.Printf("Active token: %s\n", ciServers.servers[cacheIndex].gitToken)
			base.Zlog.Infof("Active token: %s", ciServers.servers[cacheIndex].gitToken)

		}
	case "build_bios_firmware":
		if cacheIndex != -1 {
			_, tail = ShiftPath(r.URL.Path)
			keys := strings.Split(tail, "/")
			login := keys[2]
			command := keys[1]
			if !checkAccess(w, r, login, command) {
				w.Write([]byte("Access denied"))
				return
			}
			// We have to forward the request to the compile server
			// which will start the compilation process and return
			// the code to connect to the ttyd daemon
			//fmt.Printf("Forward biosfirmware build\n")
			base.Zlog.Infof("Forward bios firmware build")
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].compileIP + compileTCPPort)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].compileIP + compileTCPPort
			// This approach is not really safe we shall transfer the Token through a specific call
			r.URL.Path = tail + "/" + ciServers.servers[cacheIndex].gitToken
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "build_bmc_firmware":
		if cacheIndex != -1 {
			_, tail = ShiftPath(r.URL.Path)
			keys := strings.Split(tail, "/")
			login := keys[2]
			command := keys[1]
			if !checkAccess(w, r, login, command) {
				w.Write([]byte("Access denied"))
				return
			}
			// We have to forward the request to the compile server
			// which will start the compilation process and return
			// the code to connect to the ttyd daemon
			//fmt.Printf("Forward bmcfirmware build\n")
			base.Zlog.Infof("Forward bmc firmware build")
			url, _ := url.Parse("http://" + ciServers.servers[cacheIndex].compileIP + compileTCPPort)
			proxy := httputil.NewSingleHostReverseProxy(url)
			r.URL.Host = "http://" + ciServers.servers[cacheIndex].compileIP + compileTCPPort
			r.URL.Path = tail + "/" + ciServers.servers[cacheIndex].gitToken
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w, r)
		}
	case "load_built_smbios":
		if cacheIndex != -1 {
			// we must tell to the controller node that he needs to download the BIOS
			// from our user from the storage node and to start the em100
			_, tail = ShiftPath(r.URL.Path)
			keys := strings.Split(tail, "/")
			login := keys[2]
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET", "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/load_from_storage_smbios/"+login, nil)
			_, _ = client.Do(req)
		}
	case "load_built_openbmc":
		if cacheIndex != -1 {
			// we must tell to the controller node that he needs to download the BIOS
			// from our user from the compile node and to start the em100
			_, tail = ShiftPath(r.URL.Path)
			keys := strings.Split(tail, "/")
			login := keys[2]
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET", "http://"+ciServers.servers[cacheIndex].ip+ciServers.servers[cacheIndex].tcpPort+"/load_from_storage_bmc/"+login, nil)
			_, _ = client.Do(req)
		}
	case "":
		b, _ := ioutil.ReadFile(staticAssetsDir + "/html/homepage.html") // just pass the file name
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

func bmcweb(w http.ResponseWriter, r *http.Request) {
	// Let's print the session ID
	cookie, err := r.Cookie("osfci_cookie")

	// If the request is for a favicon.ico file we are just returning
	// we do not offer such icon currently ;)
	head, _ := ShiftPath(r.URL.Path)
	if head == "favicon.ico" {
		return
	}

	// We must validate if the user is coming with a cookie
	// if yes we must look to which server it is allocated
	// if it is not allocated to any then we probably need to reroute him
	// to the homepage !

	bmcIP := ""
	if err == nil {
		if cookie.Value != "" {
			// We must get the IP address from the cache
			for i := range ciServers.servers {
				if ciServers.servers[i].currentOwner == cookie.Value {
					if time.Now().Before(ciServers.servers[i].expiration) {
						// We still own the server and we can go to the BMC
						bmcIP = ciServers.servers[i].bmcIP
					}
				}
			}
		} else {
			if DNSDomain != "" {
				http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
			}
			return
		}
	} else {
		if DNSDomain != "" {
			http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
		}
		return
	}
	if bmcIP == "" {
		if DNSDomain != "" {
			http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
		}
		return
	}
	// We must know if iLo is started or not ?
	// if not then we have to reroute to the actual homepage
	// We can make a request to the website or
	conn, err := net.DialTimeout("tcp", bmcIP+":443", 220*time.Millisecond)
	if err != nil {
		if DNSDomain != "" {
			http.Redirect(w, r, "https://"+DNSDomain+"/ci", 302)
		}
		return
	}
	conn.Close()
	// Must specify the iLo Web address
	url, _ := url.Parse("https://" + bmcIP + ":443")
	proxy := httputil.NewSingleHostReverseProxy(url)
	var InsecureTransport http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 10 * time.Second,
	}
	// Our OpenBMC has a self signed certificate
	proxy.Transport = InsecureTransport
	// Internal gateway IP address
	// Must reroute on myself and port 443
	url, _ = url.Parse("http://" + r.Header.Get("Host"))
	r.URL.Host = "https://" + url.Hostname() + ":443/"
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	proxy.ServeHTTP(w, r)

}

//Default Intialize
func init() {

	config := base.Configuration{
		EnableConsole:     false,                                   //print output on the console, Good for debugging in local
		ConsoleLevel:      base.Debug,                              //Debug level log
		ConsoleJSONFormat: false,                                   //Console log in JSON format, false will print in raw format on console
		EnableFile:        true,                                    // Logging in File
		FileLevel:         base.Info,                               // File log level
		FileJSONFormat:    true,                                    // File JSON Format, False will print in file in raw Format
		FileLocation:      "/usr/local/production/logs/server.log", //File location where log needs to be appended
	}

	err := base.NewLogger(config)
	if err != nil {
		base.Zlog.Fatalf("Could not instantiate server log %s", err.Error())
	}
	base.Zlog.Infof("Starting server logger...")
}

func main() {
	base.Zlog.Infof("Starting server...")
	print("=============================== \n")
	print("| Starting frontend           |\n")
	print("| Development version -       |\n")
	print("| Private use only            |\n")
	print("=============================== \n")
	print(" Please do not forget to set TLS_CERT_PATH/TLS_KEY_PATH/STATIC_ASSETS_DIR to there relevant path\n")

	err := initServerconfig()
	//If there is error reading the config file log error and exit
	if err != nil {
		base.Zlog.Fatalf("Initialization error: %s", err.Error())
	}

	mux := http.NewServeMux()

	// Highest priority must be set to the signed request
	mux.HandleFunc("/ci/", home)
	mux.HandleFunc("/user/", user)
	mux.HandleFunc("/", bmcweb)

	// We must build our server pool for the moment
	// This is define by the environment variable
	// But this could be done by a registration mechanism later

	var newFamily serverProduct
	var family string
	for i := 1; ; i++ {
		family = fmt.Sprintf("family%d", i)
		viperstring := "serverfamily." + family
		if viper.IsSet(viperstring) {
			brandstring := viperstring + ".Brand"
			modelstring := viperstring + ".model"
			activestring := viperstring + ".Active"
			newFamily.Brand = viper.GetString(brandstring)
			newFamily.Product = viper.GetString(modelstring)
			newFamily.Active = viper.Get(activestring).(int)
			ciServersProducts = append(ciServersProducts, newFamily)
			continue
		} else {
			break
		}
	}

	var newEntry serverEntry
	var ctrl string
	for i := 1; ; i++ {
		ctrl = fmt.Sprintf("ctrl%d", i)
		viperstring := "controller." + ctrl
		if viper.IsSet(viperstring) {
			serverstring := viperstring + ".servername"
			bmcipstring := viperstring + ".SUTbmcIP"
			compileripstring := viperstring + ".compilerIP"
			ipstring := viperstring + ".ip"
			tcpportstring := viperstring + ".tcpPort"
			typetring := viperstring + ".SUTtype"
			newEntry.servername = viper.GetString(serverstring)
			newEntry.ip = viper.GetString(ipstring)
			newEntry.tcpPort = viper.GetString(tcpportstring)
			newEntry.compileIP = viper.GetString(compileripstring)
			newEntry.currentOwner = ""
			newEntry.gitToken = ""
			newEntry.expiration = time.Now()
			newEntry.bmcIP = viper.GetString(bmcipstring)
			newEntry.queue = 0
			servertype := viper.GetString(typetring)
			fmt.Println("servertype=", servertype)
			switch servertype {
			case "DL360_Gen10":
				newEntry.ProductIndex = 0
			case "DL325_GEN10PLUS":
				newEntry.ProductIndex = 1
			}
			ciServers.mux.Lock()
			ciServers.servers = append(ciServers.servers, newEntry)
			ciServers.mux.Unlock()
			continue
		} else {
			break
		}
	}

	if DNSDomain != "" {
		// if DNS_DOMAIN is set then we run in a production environment
		// we must get the directory where the certificates will be stored
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(certStorage),
			HostPolicy: autocert.HostWhitelist(DNSDomain),
		}

		server := &http.Server{
			Addr:         ":443",
			Handler:      mux,
			ReadTimeout:  600 * time.Second,
			WriteTimeout: 600 * time.Second,
			IdleTimeout:  120 * time.Second,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}

		go func() {
			h := certManager.HTTPHandler(nil)
			if err := http.ListenAndServe(":http", h); err != http.ErrServerClosed {
				base.Zlog.Fatalf("Server service error: %s", err.Error())
			}
		}()

		server.ListenAndServeTLS("", "")
	} else {
		go http.ListenAndServe(":80", http.HandlerFunc(httpsRedirect))
		// Launch TLS server
		if err := http.ListenAndServeTLS(":443", tlsCertPath, tlsKeyPath, mux); err != nil {
			base.Zlog.Fatalf("Server TLS error: %s", err.Error())
		}
	}
}
