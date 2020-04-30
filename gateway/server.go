package main

import (
	"net/http"
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
	"crypto/tls"
	"net"
	"time"
)


var tlsCertPath = os.Getenv("TLS_CERT_PATH")
var tlsKeyPath = os.Getenv("TLS_KEY_PATH")
var DNSDomain = os.Getenv("DNS_DOMAIN")
var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
var TTYDHostConsole = os.Getenv("TTYD_HOST_CONSOLE_PORT")
var TTYDem100Bios = os.Getenv("TTYD_EM100_BIOS_PORT")
var TTYDem100iLO = os.Getenv("TTYD_EM100_ILO_PORT")
var CTRLIp = os.Getenv("CTRL_IP")
var ExpectediLOIp = os.Getenv("EXPECT_ILO_IP")


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


func home(w http.ResponseWriter, r *http.Request) {
	head, tail := ShiftPath( r.URL.Path)
	if ( head == "ci" ) {
		head,_ = ShiftPath(tail)
	}
	switch ( head ) {
		case "console":
			fmt.Printf("Console request\n");
		        url, _ := url.Parse("http://"+CTRLIp+TTYDHostConsole)
		        proxy := httputil.NewSingleHostReverseProxy(url)
		        r.URL.Host = "http://"+CTRLIp+TTYDHostConsole
			filePath :=  strings.Split(tail,"/")
			r.URL.Path = "/"
			if ( len(filePath) > 2 ) {
				r.URL.Path = r.URL.Path + filePath[2]
			}
		        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			proxy.ServeHTTP(w , r)
		case "smbiosconsole":
                        url, _ := url.Parse("http://"+CTRLIp+TTYDem100Bios)
                        proxy := httputil.NewSingleHostReverseProxy(url)
                        r.URL.Host = "http://"+CTRLIp+TTYDem100Bios
                        filePath :=  strings.Split(tail,"/")
                        r.URL.Path = "/"
                        if ( len(filePath) > 2 ) {
                                r.URL.Path = r.URL.Path + filePath[2]
                        }
                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                        proxy.ServeHTTP(w , r)
		case "poweron":
			fmt.Printf("Poweron request\n");
			client := &http.Client{}
                        var req *http.Request
                        req, _ = http.NewRequest("GET","http://"+CTRLIp+"/poweron", nil)
                        _, _  = client.Do(req)
		case "poweroff":
			fmt.Printf("Poweroff request\n");
			client := &http.Client{}
                        var req *http.Request
                        req, _ = http.NewRequest("GET","http://"+CTRLIp+"/poweroff", nil)
                        _, _  = client.Do(req)
		case "iloconsole":
                        url, _ := url.Parse("http://"+CTRLIp+TTYDem100iLO)
                        proxy := httputil.NewSingleHostReverseProxy(url)
                        r.URL.Host = "http://"+CTRLIp+TTYDem100iLO
                        filePath :=  strings.Split(tail,"/")
                        r.URL.Path = "/"
                        if ( len(filePath) > 2 ) {
                                r.URL.Path = r.URL.Path + filePath[2]
                        }
                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                        proxy.ServeHTTP(w , r)
		case "startilo":
			// we must forward the request to the relevant test server
			client := &http.Client{}
			var req *http.Request
			req, _ = http.NewRequest("GET","http://"+CTRLIp+"/startilo", nil)
		        _, _  = client.Do(req)

			client = &http.Client{}
                        req, _ = http.NewRequest("GET","http://"+CTRLIp+"/startiloconsole", nil)
                        _, _  = client.Do(req)
		case "startsmbios":
			// we must forward the request to the relevant test server
                        client := &http.Client{}
                        var req *http.Request
                        req, _ = http.NewRequest("GET","http://"+CTRLIp+"/startsmbios", nil)
                        _, _  = client.Do(req)
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
		case "mp4":
			b, _ := ioutil.ReadFile(staticAssetsDir+tail) // just pass the file name
                        w.Header().Set("Content-Type", "video/mp4")
                        w.Write(b)
		case "ilofirmware":
			// We must forward the request
			fmt.Printf("Forward ilofirmware upload\n");
                        url, _ := url.Parse("http://"+CTRLIp)
                        proxy := httputil.NewSingleHostReverseProxy(url)
                        r.URL.Host = "http://"+CTRLIp
                        r.URL.Path = "/ilofirmware"
                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                        proxy.ServeHTTP(w , r)
		case "biosfirmware":
			// We must forward the request
                        fmt.Printf("Forward biosfirmware upload\n");
                        url, _ := url.Parse("http://"+CTRLIp)
                        proxy := httputil.NewSingleHostReverseProxy(url)
                        r.URL.Host = "http://"+CTRLIp
                        r.URL.Path = "/biosfirmware"
                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                        proxy.ServeHTTP(w , r)
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

func iloweb(w http.ResponseWriter, r *http.Request){
	// We must know if iLo is started or not ?
	// if not then we have to reroute to the actual homepage
	fmt.Printf("Contacting iLo\n")
	// We can make a request to the website or
	conn, err := net.DialTimeout("tcp", ExpectediLOIp+":443", 220*time.Millisecond)
	if ( err != nil ) {
		if ( DNSDomain != "" ) {
			http.Redirect(w, r, "https://"+DNSDomain+"/ci", 301)
		}
		return
	} else {
		conn.Close()
	}
	// Must specify the iLo Web address
	fmt.Printf("iLo is answering - Forwarding the request\n")
	url, _ := url.Parse("https://"+ExpectediLOIp+":443")
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
        url, _ := url.Parse("http://"+r.Header.Get("Host"))
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
    mux.HandleFunc("/",iloweb)
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
