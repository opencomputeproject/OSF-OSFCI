package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/linuxboot/contest/cmds/clients/contestcli/cli"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"net/http"
	"strings"
	"time"
	"regexp"
)

func main() {
	addr := flag.String("addr", "http://localhost:8080", "Contest server URL")
	user := flag.String("user", "", "Username")
	testlist := flag.String("tests", "", "List of testcases")
	logpath := flag.String("log", "/tmp/", "Log Path")
	flag.Parse()
	if len(*user) == 0 {
		fmt.Println("Invalid inputs")
		return
	}
	timestamp := time.Now().Format("20060102150405")
	logFile := path.Join(*logpath, "contest_"+*user+"_"+timestamp+".log")
	loghandler, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, loghandler)
	log.SetOutput(mw)
	userDir := path.Join(*logpath, "contest_"+*user+"_"+timestamp)
	start(*addr, *user, *testlist, userDir)
	log.Println("Done")
}

func start(addr string, user string, testlist string, userDir string) {
	tests := strings.Split(testlist, ",")
	if len(tests) == 0 {
		log.Println("Please provide the tests names")
		return
	}
	zipArchive, _ := os.Create(userDir + ".zip")
	defer zipArchive.Close()
	writer := zip.NewWriter(zipArchive)
	log.Printf("User:%s, Tests:%s\n", user, testlist)
	for _, test := range tests {
		if _, err := os.Stat(test); os.IsNotExist(err) {
			log.Printf("File [%s] does not exist\n", test)
			continue
		}
		log.Printf("Executing the test: %s\n", test)
		testname := getTestName(test)
		outfile := path.Join(userDir, "contest_"+testname+"_output.log")
		outhandler, _ := os.OpenFile(outfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		defer outhandler.Close()
		var out bytes.Buffer
		input := []string{os.Args[0], "--addr", addr, "start", test}
		if err := cli.CLIMain(input[0], input[1:], &out); err != nil {
			log.Printf("%v\n", err)
		}
		var jobData map[string]interface{}
		json.Unmarshal(out.Bytes(), &jobData)
		log.Println(out.String())
		zipwriter, _ := writer.Create(testname + "_console_output.json")
		if _, ok := jobData["Data"]; ok {
			jobID := fmt.Sprintf("%d", int(jobData["Data"].(map[string]interface{})["JobID"].(float64)))
			log.Printf("Job started successfully, Job ID: %s\n", jobID)
			log.Printf("\nWaiting for job to complete\n")
			time.Sleep(30 * time.Second)
			out, err := status(addr, user, jobID)
			if err != nil {
				outhandler.WriteString(err.Error())
				io.WriteString(zipwriter, err.Error())
			} else {
				outhandler.WriteString(out.String())
				io.WriteString(zipwriter, out.String())
			}
			log.Printf("Downloading the test logs from contest server")
			reg := regexp.MustCompile(`(.+:)(\d+)`)
			download_url := reg.ReplaceAllString(addr, "${1}8789") 
			download_url = fmt.Sprintf("%s/logs/%s",download_url, jobID)
			resp, _ := http.Get(download_url)
			defer resp.Body.Close()
			log.Printf("Log download status:", resp.Status)
			if resp.StatusCode != 200 {
				log.Printf("Unable to download")
			} else {
				filename := fmt.Sprintf("testlogs_%s.zip", jobID)
				out, _ := writer.Create(filename)
				io.Copy(out, resp.Body)
			}
		} else {
			log.Println("Error: Unable to execute the testcase")
			outhandler.WriteString("Error: Unable to execute the testcase")
			io.WriteString(zipwriter, "Error: Unable to execute the testcase")
		}
	}
	writer.Close()
}

func status(addr string, user string, jobID string) (bytes.Buffer, error) {
	var out bytes.Buffer
	for attempt := 1; attempt <= 20; attempt++ {	
		input := []string{os.Args[0], "--addr", addr, "status", jobID}
		if err := cli.CLIMain(input[0], input[1:], &out); err != nil {
			log.Printf("%v\n", err)
			return out, err
		}
		var JobCompletionEvents = map[string]int{
			"JobStateCompleted"	: 0, 
			"JobStateFailed"	: 0, 
			"JobStateCancelled"	: 0, 
			"JobStateCancellationFailed" : 0,
		}
		var jobStatus map[string]interface{}
		json.Unmarshal(out.Bytes(), &jobStatus)
		status := jobStatus["Data"].(map[string]interface{})["Status"].(map[string]interface{})["State"].(string)
		log.Println("Job status: ", status)
		if _, ok := JobCompletionEvents[status]; !ok {
			log.Println("Job is still running")
			time.Sleep(30 * time.Second)
			out.Reset()
		} else { 
			log.Println("Job is completed")
			log.Println(out.String())	
			break
		}
	}
	return out, nil
}

func getTestName(testpath string) string {
	return strings.TrimSuffix(path.Base(testpath), filepath.Ext(path.Base(testpath)))
}
