package main

import (
	"base/base"
	"encoding/base64"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"path/filepath"
)

var storageRoot string

// write operation must be protected by a Mutex
var file sync.RWMutex

//Initialize storage config
func initStorageconfig() error {
	viper.SetConfigName("gatewayconf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/production/config/")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	storageRoot = viper.GetString("STORAGE_ROOT")

	return nil
}

// This is getting a user file entry

func getEntry(username string) (string, int) {
	// The first letter of the username is used as a directory entry
	// if the directory exist we check for the usenarme.conf entry into it
	// if it is there we return the content of the file
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
	if !os.IsNotExist(err) {
		// The directory exist we must now check if the file exist
		_, err := os.Stat(storageRoot + "/" + string(username[0]) + "/" + username)
		if !os.IsNotExist(err) {
			// We must return the file content into a string
			b, _ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + username)
			return string(b), 1
		}
		return "", 0
	}
	return "", 0
}

// This is creating a user file entry

func createEntry(username string, content string) int {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
	file.Lock()
	defer file.Unlock()
	if os.IsNotExist(err) {
		// we must create the directory which will contain the file
		_ = os.Mkdir(storageRoot+"/"+string(username[0]), os.ModePerm)
	}
	base.Zlog.Infof("Saving the data for user: %s", username)
	_ = ioutil.WriteFile(storageRoot+"/"+string(username[0])+"/"+username, []byte(content), os.ModePerm)
	return 1
}

func createImage(username string, content string) int {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))

	file.Lock()
	defer file.Unlock()
	if os.IsNotExist(err) {
		// we must create the directory which will contain the file
		_ = os.Mkdir(storageRoot+"/"+string(username[0]), os.ModePerm)
	}
	// We have to remove the "base64, stuff"
	coI := strings.Index(content, ",")
	rawImage := string(content)[coI+1:]
	decodedBody, _ := base64.StdEncoding.DecodeString(rawImage)
	_ = ioutil.WriteFile(storageRoot+"/"+string(username[0])+"/"+username+".jpg", []byte(decodedBody), os.ModePerm)
	return 1
}

func storeFirmware(username string, r *http.Request, firmware string) int {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))

	file.Lock()
	defer file.Unlock()
	if os.IsNotExist(err) {
		// we must create the directory which will contain the file
		_ = os.Mkdir(storageRoot+"/"+string(username[0]), os.ModePerm)
	}
	_ = ioutil.WriteFile(storageRoot+"/"+string(username[0])+"/"+firmware+"_"+username+".rom", base.HTTPGetBody(r), os.ModePerm)
	return 1
}

func storeLog(username string, r *http.Request, firmware string) int {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))

	file.Lock()
	defer file.Unlock()
	if os.IsNotExist(err) {
		// we must create the directory which will contain the file
		_ = os.Mkdir(storageRoot+"/"+string(username[0]), os.ModePerm)
	}
	_ = ioutil.WriteFile(storageRoot+"/"+string(username[0])+"/"+firmware+"_"+username+".log", base.HTTPGetBody(r), os.ModePerm)
	return 1
}

func getSystemBIOS(username string, w http.ResponseWriter, recipe string) {
	content, _ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + "linuxboot_" + recipe + "_" + username + ".rom")
	w.Header().Add("Content-Length", strconv.Itoa(len(content)))
	w.Write(content)
}

func getSystemBIOSBuildLog(username string, w http.ResponseWriter, recipe string) {
	content, _ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + "linuxboot_" + recipe + "_" + username + ".log")
	w.Header().Add("Content-Length", strconv.Itoa(len(content)))
	w.Write(content)
}

func getOpenBMC(username string, w http.ResponseWriter, recipe string) {
	content, _ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + "openbmc_" + recipe + "_" + username + ".rom")
	w.Header().Add("Content-Length", strconv.Itoa(len(content)))
	w.Write(content)
}

func getOpenBMCBuildLog(username string, w http.ResponseWriter, recipe string) {
	content, _ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + "openbmc_" + recipe + "_" + username + ".log")
	w.Header().Add("Content-Length", strconv.Itoa(len(content)))
	w.Write(content)
}

func getImage(username string) string {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
	base.Zlog.Infof("Get the image: %s: %s", err, username)
	file.Lock()
	defer file.Unlock()
	if os.IsNotExist(err) {
		// we must create the directory which will contain the file
		_ = os.Mkdir(storageRoot+"/"+string(username[0]), os.ModePerm)
		return ""
	}

	_, err = os.Stat(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	base.Zlog.Infof("Image: %s", err)
	if os.IsNotExist(err) {
		var staticAssetsDir = viper.GetString("STATIC_ASSETS_DIR")
		content, _ := ioutil.ReadFile(staticAssetsDir + "images/forklift.png")
		encodedContent := base64.StdEncoding.EncodeToString(content)
		return encodedContent
	}
	base.Zlog.Infof("storageRoot: %s:", storageRoot)
	content, _ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	encodedContent := base64.StdEncoding.EncodeToString(content)
	return encodedContent
}

func deleteEntry(username string, content string) int {
	_, err := os.Stat(storageRoot + "/" + string(username[0]) + "/" + username)
	base.Zlog.Infof("Delete: %s: %s", err, username)
	file.Lock()
	defer file.Unlock()
	base.Zlog.Infof("Checking if the file exists")
	if !os.IsNotExist(err) {
		base.Zlog.Infof("deleting user file")
		//_ = os.Remove(storageRoot + "/" + string(username[0]) + "/" + username)
	}
	_, err = os.Stat(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	if !os.IsNotExist(err) {
		base.Zlog.Infof("deleting user image")
		//_ = os.Remove(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	}
	return 1
}

func deleteUserData(username string, content string) int {
	base.Zlog.Infof("Deleting the User data of : %s", username)
	findDeleteUserData(storageRoot + "/" + string(username[0]) + "/linuxboot_*" + username + ".rom")
	findDeleteUserData(storageRoot + "/" + string(username[0]) + "/linuxboot_*" + username + ".log")
	findDeleteUserData(storageRoot + "/" + string(username[0]) + "/openbmc_*" + username + ".rom")
	findDeleteUserData(storageRoot + "/" + string(username[0]) + "/openbmc_*" + username + ".log")
	return 1
}

func findDeleteUserData(pattern string) int {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		base.Zlog.Fatalf(err.Error())
		return 0
	}
	base.Zlog.Infof("Total number of data files found: %d", len(matches))
	for _, file := range matches {
		base.Zlog.Infof("Deleting the file: %s", file)
		//_ = os.Remove(file)
	}
	return 1
}


func distrosCallback(w http.ResponseWriter, r *http.Request) {
	// We must breakdown the words, because directory filename is the last word
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "401 Malformed URI", 401)
	}
	if path[2] == "" {
		// We must provide the directory content from distros
		files, _ := ioutil.ReadDir(storageRoot + "/distros")
		var answer string
		var count int
		if len(files) > 0 {
			answer = "{ \"files\": ["
			count = 0
			for _, file := range files {
				if count == 1 {
					answer = answer + ","
				}
				answer = answer + "\"" + file.Name() + "\""
				count = 1
			}
			answer = answer + "] }"
		}
		w.Write([]byte(answer))
	} else {
		// We must serve the file
		http.ServeFile(w, r, storageRoot+"/distros/"+path[2])
	}
}

func userCallback(w http.ResponseWriter, r *http.Request) {
	var username string
	var filecontent string
	var returnValue int
	// We must breakdown the words, because username is not always the last word
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "401 Malformed URI", 401)
		return
	}
	username = path[2]
	var command string
	var recipe string
	base.Zlog.Infof("path: %s", r.URL.Path)
	if len(path) > 3 {
	        base.Zlog.Infof("path 3: %s", path[3])
		command = path[3]
	}
	if len(path) > 4 {
		recipe = path[4]
	}
	switch r.Method {
	case http.MethodGet:
		// Serve the resource.
		// I must return the content of the user file if it does exist otherwise
		// an error
		switch command {
		case "avatar":
			w.Write([]byte(getImage(username)))
		case "get_firmware":
			getSystemBIOS(username, w, recipe)
		case "get_bmc_firmware":
			getOpenBMC(username, w, recipe)
		case "get_firmware_build_log":
			getSystemBIOSBuildLog(username, w, recipe)
		case "get_bmc_firmware_build_log":
			getOpenBMCBuildLog(username, w, recipe)
		default:
			filecontent, returnValue = getEntry(username)
			if returnValue != 0 {
				fmt.Fprint(w, filecontent)
			} else {
				fmt.Fprintf(w, "Error")
			}
		}
	case http.MethodPut:
		// Update an existing record.
		if r.Header.Get("Content-Type") != "image/jpg" {
			if r.Header.Get("Content-Type") == "application/octet-stream" {
				// We got a firmware
				if command == "linuxboot" {
					storeFirmware(username, r, "linuxboot_"+recipe)
				} else {
					if command == "openbmc" {
						storeFirmware(username, r, "openbmc_"+recipe)
					}
				}
			} else {
				if r.Header.Get("Content-Type") == "text/plain" {
					if command == "linuxboot" {
						storeLog(username, r, "linuxboot_"+recipe)
					} else {
						if command == "openbmc" {
							storeLog(username, r, "openbmc_"+recipe)
						}
					}
				} else {
					base.Zlog.Infof("For Delete: : %s", username)
					createEntry(username, string(base.HTTPGetBody(r)))
				}
			}
		} else {
			createImage(username, string(base.HTTPGetBody(r)))
		}
	case http.MethodDelete:
		switch command {
			case "delete_user_data":
				base.Zlog.Infof("Deleting the data of user: %s", username)
				deleteUserData(username, string(base.HTTPGetBody(r)))
			default:
				base.Zlog.Infof("Deleting the user: %s", username)
				deleteEntry(username, string(base.HTTPGetBody(r)))
				base.Zlog.Infof("Deleted the user: %s", username)
		}
	default:
	}
}

//Default Intialize
func init() {

	config := base.Configuration{
		EnableConsole:     false,                                    //print output on the console, Good for debugging in local
		ConsoleLevel:      base.Debug,                               //Debug level log
		ConsoleJSONFormat: false,                                    //Console log in JSON format, false will print in raw format on console
		EnableFile:        true,                                     // Logging in File
		FileLevel:         base.Info,                                // File log level
		FileJSONFormat:    true,                                     // File JSON Format, False will print in file in raw Format
		FileLocation:      "/usr/local/production/logs/storage.log", //File location where log needs to be appended
	}

	err := base.NewLogger(config)
	if err != nil {
		base.Zlog.Fatalf("Could not instantiate storage log %s", err.Error())
	}
	base.Zlog.Infof("Starting storage logger...")
}

func main() {
	base.Zlog.Infof("Starting storage backend...")
	print("=============================== \n")
	print("| Starting storage backend    |\n")
	print("| Development version -       |\n")
	print("| Private use only            |\n")
	print("=============================== \n")

	err := initStorageconfig()
	if err != nil {
		base.Zlog.Fatalf("Storage config initialization error: %s", err.Error())
	}

	mux := http.NewServeMux()
	var StorageURI = viper.GetString("STORAGE_URI")
	var StorageTCPPORT = viper.GetString("STORAGE_TCPPORT")

	fmt.Println("StorageURI =", StorageURI)
	fmt.Println("StorageTCPPORT =", StorageTCPPORT)
	mux.HandleFunc("/user/", userCallback)
	mux.HandleFunc("/distros/", distrosCallback)

	if err := http.ListenAndServe(StorageURI+StorageTCPPORT, mux); err != nil {
		base.Zlog.Fatalf("Storage service error: %s", err.Error())
	}
}
