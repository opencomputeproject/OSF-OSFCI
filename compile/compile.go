package main

import (
	"net/http"
	"strings"
	"path"
	"log"
)

func ShiftPath(p string) (head, tail string) {
    p = path.Clean("/" + p)
    i := strings.Index(p[1:], "/") + 1
    if i <= 0 {
        return p[1:], "/"
    }
    return p[1:i], p[i:]
}


func home(w http.ResponseWriter, r *http.Request) {
	head,_ := ShiftPath( r.URL.Path)
	switch ( head ) {
		case "buildilofirmware":
                        switch r.Method {
                                case http.MethodPost:
				}
		case "buildbiosfirmware":
			switch r.Method {
                                case http.MethodPost:
                        }
		default:
	}
}

func main() {
    print("=============================== \n")
    print("| Starting frontend           |\n")
    print("| Development version -       |\n")
    print("=============================== \n")


    mux := http.NewServeMux()

    // Highest priority must be set to the signed request
    mux.HandleFunc("/",home)


    log.Fatal(http.ListenAndServe(":80", mux))
}
