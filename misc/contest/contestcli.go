package main

import (
        "fmt"
        "os"
	"io"
	"flag"
	"log"
	"strings"
	"bytes"
	"time"
	"path"
	"encoding/json"
	"github.com/linuxboot/contest/cmds/clients/contestcli/cli"
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
	logFile := path.Join(*logpath, "contest_" + *user + ".log")
	loghandler, err := os.OpenFile(logFile, os.O_CREATE | os.O_APPEND | os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, loghandler)
	log.SetOutput(mw)
	start(*addr, *user, *testlist)
	log.Println("Done")
}

func start(addr string, user string, testlist string){
	tests := strings.Split(testlist, ",")
	if len(tests) == 0 {
		log.Println("Please provide the tests names")
		return 
	}
	fmt.Printf("User:%s, Tests:%s\n", user, testlist)
	for _, test := range(tests) {
		if _, err := os.Stat(test); os.IsNotExist(err){
			log.Printf("File [%s] does not exist\n", test);
			continue
		}
		log.Printf("Executing the test: %s\n", test)
		var out bytes.Buffer
		input := []string{os.Args[0], "--addr", addr, "start", test}
		if err := cli.CLIMain(input[0], input[1:], &out); err != nil {
			log.Printf("%v\n", err)
		}
		var jobData map[string]interface{}
		json.Unmarshal(out.Bytes(), &jobData)
		log.Println(out.String())
		if _, ok := jobData["Data"]; ok {
			jobID := fmt.Sprintf("%d", int(jobData["Data"].(map[string]interface{})["JobID"].(float64)))
			log.Printf("Job started succesfully, Job ID: %s\n", jobID)
			log.Printf("\nWaiting for job to complete\n")
			time.Sleep(5 * time.Second)
			status(addr, user, jobID)
		}else{
			log.Println("Error: Unable to execute the testcase")
		}
	}
}

func status(addr string, user string, jobID string){
	var out bytes.Buffer
	input := []string{os.Args[0], "--addr", addr, "status", jobID}
	if err := cli.CLIMain(input[0], input[1:], &out); err != nil {
		log.Printf("%v\n", err)
	}
	var jobData map[string]interface{}
	json.Unmarshal(out.Bytes(), &jobData)
	log.Println(out.String())
}

