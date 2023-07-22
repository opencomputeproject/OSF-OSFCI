package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Go utility to fetch logs from the contest server
// Its running on port 8789 on contest server

var CONTEST_LOGDIR string

func main() {
	f, err := os.OpenFile("contest_log_server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("Starting the Content log server")
	if dir, exists := os.LookupEnv("CONTEST_LOGDIR"); exists {
		CONTEST_LOGDIR = dir
		if _, err := os.Stat(CONTEST_LOGDIR); os.IsNotExist(err) {
			if err := os.MkdirAll(CONTEST_LOGDIR, os.ModePerm); err != nil {
				log.Fatalf("failed to create proc: %w", err)
			}
		}
	} else {
		CONTEST_LOGDIR, err = os.UserHomeDir()
		if err != nil {
			log.Fatalf("failed to create proc: %w", err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/logs/", home)
	if err := http.ListenAndServe(":8789", mux); err != nil {
		log.Fatalf("Contest log service error: %w", err)
	}
}

func home(w http.ResponseWriter, req *http.Request) {
	_, tail := shiftPath(req.URL.Path)
	log.Println("URL:", tail)
	url := strings.Split(tail, "/")
	log.Println(url[1])
	if len(url[1]) == 0 || len(url) > 2 {
		http.NotFound(w, req)
		return
	}
	logdir := path.Clean(CONTEST_LOGDIR + "/" + url[1])
	log.Println(logdir)
	if _, err := os.Stat(logdir); os.IsNotExist(err) {
		http.NotFound(w, req)
		return
	}

	writer := zip.NewWriter(w)
	defer writer.Close()
	err := filepath.Walk(logdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		f, err := writer.Create(info.Name())
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", "contest_testlog_"+url[1]))
	return
}

// ShiftPath cleans up path
func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
