// OSFCI Compiler module used to provide build-compile routines fo

package main

import (
	"base/base"
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

var compileTCPPort string
var startLinuxbootBuildBin string
var startOpenBMCBuildBin string
var binariesPath string
var firmwaresPath string
var storageURI string
var storageTCPPort string

// OpenBMCCommand  initialized
var OpenBMCCommand *exec.Cmd = nil

// OpenBMCOutput type such that it can be managed
var OpenBMCOutput io.ReadCloser

// LinuxBOOTCommand  initialized
var LinuxBOOTCommand *exec.Cmd = nil

// LinuxBOOTOutput type such that it can be managed
var LinuxBOOTOutput io.ReadCloser

// OpenBMCBuildChannel chan setup
var OpenBMCBuildChannel chan string

// LinuxBOOTBuildChannel chan setup
var LinuxBOOTBuildChannel chan string
var dockerClient *client.Client
var username string
var linuxbootDockerID string
var openbmcDockerID string

// initialize compiler config
func initCompilerconfig() error {
	viper.SetConfigName("compiler1conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/production/config/")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	compileTCPPort = viper.GetString("COMPILE_TCPPORT")
	startLinuxbootBuildBin = viper.GetString("LINUXBOOT_BUILD")
	startOpenBMCBuildBin = viper.GetString("OPENBMC_BUILD")
	binariesPath = viper.GetString("BINARIES_PATH")
	firmwaresPath = viper.GetString("FIRMWARES_PATH")
	storageURI = viper.GetString("STORAGE_URI")
	storageTCPPort = viper.GetString("STORAGE_TCPPORT")

	return nil
}

// ShiftPath to check if a docker container is running
// docker inspect -f '{{.State.Running}}' linuxboot_vejmarie2
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func containerList() []types.Container {
	var options types.ContainerListOptions
	options.All = true
	containers, err := dockerClient.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}
	return containers
}

func isRunning(prefix string) bool {
	containers := containerList()
	myUniqueID := md5.Sum([]byte(username + "\n"))
	containerName := prefix + "_" + hex.EncodeToString(myUniqueID[:])
	for _, container := range containers {
		if container.Names[0] == "/"+containerName {
			return true
		}
	}
	return false
}

func home(w http.ResponseWriter, r *http.Request) {
	head, tail := ShiftPath(r.URL.Path)
	switch head {
	case "clean_up":
		device := tail[1:]
		if len(device) > 1 {
			fmt.Printf("Device: %d\n", device)
			if device == "bmc" {
				if OpenBMCCommand != nil {
					unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
					_ = <-OpenBMCBuildChannel
					OpenBMCCommand = nil
				}
			} else {
				if device == "rom" {
					if LinuxBOOTCommand != nil {
						unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
						_ = <-LinuxBOOTBuildChannel
						LinuxBOOTCommand = nil
					}
				}
			}
		}

		if OpenBMCCommand != nil {
			unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
			_ = <-OpenBMCBuildChannel
			OpenBMCCommand = nil
		}
		if LinuxBOOTCommand != nil {
			unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
			_ = <-LinuxBOOTBuildChannel
			LinuxBOOTCommand = nil
		}
	case "is_running":
		command := tail[1:]
		if isRunning(command) {
			w.Write([]byte("{ \"status\" : \"1\" }"))
		} else {
			w.Write([]byte("{ \"status\" : \"0\" }"))
		}
	case "get_firmware":
		login := tail[1:]
		// We must retrieve the username BIOS and return it as the response body
		if LinuxBOOTCommand != nil {
			unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
			_ = <-LinuxBOOTBuildChannel
			LinuxBOOTCommand = nil
		}
		f, _ := os.Open(firmwaresPath + "/test_" + login + ".rom")
		defer f.Close()
		firmware := make([]byte, 64*1024*1024)
		_, _ = f.Read(firmware)
		w.Write(firmware)
	case "get_bmc_firmware":
		login := tail[1:]
		// We must retrieve the username BIOS and return it as the response body
		if OpenBMCCommand != nil {
			unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
			_ = <-OpenBMCBuildChannel
			OpenBMCCommand = nil
		}
		f, _ := os.Open(firmwaresPath + "/test_openbmc_" + login + ".mtd")
		defer f.Close()
		firmware := make([]byte, 32*1024*1024)
		_, _ = f.Read(firmware)
		w.Write(firmware)
	case "build_bmc_firmware":
		switch r.Method {
		case http.MethodPut:
			var gitToken string
			if OpenBMCCommand != nil {
				unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
				_ = <-OpenBMCBuildChannel
				OpenBMCCommand = nil
			}
			fmt.Printf("Tail: %s\n", tail)
			keys := strings.Split(tail, "/")

			gitToken = "OSFCIemptyOSFCI"
			if len(keys) > 2 {
				username = keys[1]
				gitToken = keys[2]
			} else {
				username = keys[1]
			}
			fmt.Printf("%s %s\n", username, keys)
			fmt.Printf("GitToken: %s\n", gitToken)

			data := base.HTTPGetBody(r)
			keywords := strings.Fields(string(data))
			githubRepo := keywords[0]
			githubBranch := keywords[1]
			recipes := keywords[2]
			interactive := keywords[3]
			proxy := viper.GetString("PROXY")
			// We have to fork the build
			// The script is startLinuxbootBuild
			// It is getting 3 parameters
			// 1 - Username
			// 2 - Github repo address
			// 3 - Branch
			// 4 - Boards (which is a directory contained into the github repo)
			// The github repo must have a format which is
			// Second parameter shall be a string array

			var args []string
			args = append(args, username)
			args = append(args, githubRepo)
			args = append(args, githubBranch)
			args = append(args, recipes)
			args = append(args, storageURI)
			args = append(args, storageTCPPort)
			args = append(args, interactive)
			args = append(args, gitToken)

			args = append(args, proxy)
			OpenBMCCommand = exec.Command(startOpenBMCBuildBin, args...)
			OpenBMCCommand.SysProcAttr = &unix.SysProcAttr{
				Setsid: true,
			}
			if interactive != "1" {
				OpenBMCOutput, _ = OpenBMCCommand.StdoutPipe()
				OpenBMCCommand.Stderr = OpenBMCCommand.Stdout
			}
			err := OpenBMCCommand.Start()
			if err == nil {
				go func() {
					OpenBMCCommand.Wait()
					OpenBMCBuildChannel <- "done"
				}()
				if interactive == "1" {
					// We must hang off after being sure that the console daemon is properly starter
					conn, err := net.DialTimeout("tcp", "localhost:7682", 220*time.Millisecond)
					maxLoop := 5
					for err != nil && maxLoop > 0 {
						conn, err = net.DialTimeout("tcp", "localhost:7682", 220*time.Millisecond)
						maxLoop = maxLoop - 1
					}
					if err != nil {
						// Daemon has not started
						// Let's report an error
						w.Write([]byte("Error"))
						return
					}
					conn.Close()
				} else {
					scanner := bufio.NewScanner(OpenBMCOutput)
					buffer := make([]byte, 0, 64*1024)
					scanner.Buffer(buffer, 64*1024*1024)
					scanner.Scan()
					openbmcDockerID = scanner.Text()
					fmt.Printf("New container: %s\n", openbmcDockerID)
					go func() {
						var localLog []byte
						for scanner.Scan() {
							localLog = append(localLog, scanner.Bytes()...)
						}
						base.HTTPPutRequest("http://"+storageURI+storageTCPPort+"/user/"+username+"/openbmc/"+recipes+"/", []byte(localLog), "text/plain")
					}()
					// we have to push the log to the storage area
				}

			} else {
				OpenBMCCommand = nil
			}

		}
	case "build_bios_firmware":
		switch r.Method {
		case http.MethodPut:
			var gitToken string
			if LinuxBOOTCommand != nil {
				unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
				_ = <-LinuxBOOTBuildChannel
				LinuxBOOTCommand = nil
			}
			// We must retrieve the Token
			keys := strings.Split(tail, "/")

			gitToken = "OSFCIemptyOSFCI"
			if len(keys) > 2 {
				username = keys[1]
				gitToken = keys[2]
			} else {
				username = keys[1]
			}
			base.Zlog.Infof("%s %s", username, keys)
			base.Zlog.Infof("GitToken: %s", gitToken)

			data := base.HTTPGetBody(r)
			keywords := strings.Fields(string(data))
			githubRepo := keywords[0]
			githubBranch := keywords[1]
			board := keywords[2]
			interactive := keywords[3]
			proxy := viper.GetString("PROXY")
			// We have to fork the build
			// The script is startLinuxbootBuild
			// It is getting 3 parameters
			// 1 - Username
			// 2 - Github repo address
			// 3 - Branch
			// 4 - Boards (which is a directory contained into the github repo)
			// The github repo must have a format which is
			// Second parameter shall be a string array

			var args []string
			args = append(args, username)
			args = append(args, githubRepo)
			args = append(args, githubBranch)
			args = append(args, board)
			args = append(args, storageURI)
			args = append(args, storageTCPPort)
			args = append(args, interactive)
			args = append(args, gitToken)

			args = append(args, proxy)

			for i := 0; i < len(args); i++ {
				print(args[i] + "\n")
			}

			LinuxBOOTCommand = exec.Command(startLinuxbootBuildBin, args...)
			if interactive != "1" {
				LinuxBOOTOutput, _ = LinuxBOOTCommand.StdoutPipe()
				LinuxBOOTCommand.Stderr = LinuxBOOTCommand.Stdout
			}
			err := LinuxBOOTCommand.Start()
			if err == nil {
				go func() {
					LinuxBOOTCommand.Wait()
					LinuxBOOTBuildChannel <- "done"
				}()
				if interactive == "1" {
					// We must hang off after being sure that the console daemon is properly starter
					conn, err := net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
					maxLoop := 5
					for err != nil && maxLoop > 0 {
						conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
						maxLoop = maxLoop - 1
					}
					if err != nil {
						// Daemon has not started
						// Let's report an error
						w.Write([]byte("Error"))
						return
					}
					conn.Close()
				} else {
					scanner := bufio.NewScanner(LinuxBOOTOutput)
					scanner.Scan()
					linuxbootDockerID = scanner.Text()
					fmt.Printf("New container: %s\n", linuxbootDockerID)
					go func() {
						var linuxbootLog strings.Builder
						for scanner.Scan() {
							line := scanner.Text()
							linuxbootLog.WriteString(line + "\n")
						}
						recipe := strings.Split(board, "/")
						brecipe := recipe[1]

						base.HTTPPutRequest("http://"+storageURI+storageTCPPort+"/user/"+username+"/linuxboot/"+brecipe+"/", []byte(linuxbootLog.String()), "text/plain")
					}()
				}

			} else {
				LinuxBOOTCommand = nil
			}
		}
	default:
	}
}

// Default Intialize
func init() {

	config := base.Configuration{
		EnableConsole:     false,                                    //print output on the console, Good for debugging in local
		ConsoleLevel:      base.Debug,                               //Debug level log
		ConsoleJSONFormat: false,                                    //Console log in JSON format, false will print in raw format on console
		EnableFile:        true,                                     // Logging in File
		FileLevel:         base.Info,                                // File log level
		FileJSONFormat:    true,                                     // File JSON Format, False will print in file in raw Format
		FileLocation:      "/usr/local/production/logs/compile.log", //File location where log needs to be appended
	}

	err := base.NewLogger(config)
	if err != nil {
		base.Zlog.Fatalf("Could not instantiate compiler log %s", err.Error())
	}
	base.Zlog.Infof("Starting compiler logger...")
}

func main() {
	base.Zlog.Infof("Starting compiler backend...")

	err := initCompilerconfig()
	if err != nil {
		base.Zlog.Fatalf("Compiler config initialization error: %s", err.Error())
	}

	dockerClient, _ = client.NewEnvClient()

	OpenBMCBuildChannel = make(chan string)
	LinuxBOOTBuildChannel = make(chan string)
	mux := http.NewServeMux()

	// Highest priority must be set to the signed request
	mux.HandleFunc("/", home)
	if err := http.ListenAndServe(compileTCPPort, mux); err != nil {
		base.Zlog.Fatalf("Compiler service error: %s", err.Error())
	}
}
