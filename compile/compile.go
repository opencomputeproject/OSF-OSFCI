package main

import (
	"net/http"
	"strings"
	"path"
	"log"
	"net"
	"time"
        "os"
	"os/exec"
	"base"
	"fmt"
	"context"
	"golang.org/x/sys/unix"
        "github.com/docker/docker/api/types"
        "github.com/docker/docker/client"
        "crypto/md5"
        "encoding/hex"
)

var compileTcpPort = os.Getenv("COMPILE_TCPPORT")
var startLinuxbootBuildBin = os.Getenv("LINUXBOOT_BUILD")
var startOpenBMCBuildBin = os.Getenv("OPENBMC_BUILD")
var binariesPath = os.Getenv("BINARIES_PATH")
var firmwaresPath = os.Getenv("FIRMWARES_PATH")
var storageUri = os.Getenv("STORAGE_URI")
var storageTcpPort= os.Getenv("STORAGE_TCPPORT")
var OpenBMCCommand *exec.Cmd = nil
var LinuxBOOTCommand *exec.Cmd = nil
var OpenBMCBuildChannel chan string
var LinuxBOOTBuildChannel chan string
var dockerClient *client.Client
var username string

// to check if a docker container is running
// docker inspect -f '{{.State.Running}}' linuxboot_vejmarie2

func ShiftPath(p string) (head, tail string) {
    p = path.Clean("/" + p)
    i := strings.Index(p[1:], "/") + 1
    if i <= 0 {
        return p[1:], "/"
    }
    return p[1:i], p[i:]
}

func containerList() ([]types.Container) {
        containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
        if err != nil {
                panic(err)
        }
        return containers
}

func isRunning(prefix string) (bool) {
        containers := containerList()
        myUniqueId := md5.Sum([]byte(username+"\n"))
        containerName := prefix + "_" + hex.EncodeToString(myUniqueId[:])
        for _, container := range containers {
                if ( container.Names[0] == "/"+containerName ) {
                        return true
                }
        }
        return false
}

func home(w http.ResponseWriter, r *http.Request) {
	head,tail := ShiftPath( r.URL.Path)
	switch ( head ) {
		case "cleanUp":
			device := tail[1:]
			if ( len(device) > 1 ) {
				fmt.Printf("Device: %d\n", device)
				if ( device == "bmc" ) {
					unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
	                                _ = <- OpenBMCBuildChannel
	                                OpenBMCCommand = nil
				} else {
					if ( device == "rom" ) {
						unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
		                                _ = <- LinuxBOOTBuildChannel
		                                LinuxBOOTCommand = nil
					}
				}
			}
			
			if ( OpenBMCCommand != nil ) {
                                unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
                                _ = <- OpenBMCBuildChannel
				OpenBMCCommand = nil
                        }
                        if ( LinuxBOOTCommand != nil ) {
                                unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
				_ = <- LinuxBOOTBuildChannel
				LinuxBOOTCommand = nil
                        }
 		case "isRunning":
			command := tail[1:]
			if ( command == "openbmc" ) {
				if ( OpenBMCCommand != nil ) {
					w.Write([]byte("{ \"status\" : \"1\" }"))
				} else {
					w.Write([]byte("{ \"status\" : \"0\" }"))
				}
			} else {
				if ( command == "linuxboot" ) {
					if ( LinuxBOOTCommand != nil ) {
						w.Write([]byte("{ \"status\" : \"1\" }"))
					} else {
						w.Write([]byte("{ \"status\" : \"0\" }"))
					}
				}
			}
		case "getFirmware":
			login := tail[1:]
			// We must retreive the username BIOS and return it as the response body
                        if ( LinuxBOOTCommand != nil ) {
				unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
				_ = <- LinuxBOOTBuildChannel
				LinuxBOOTCommand = nil
                        }
			f, _ := os.Open(firmwaresPath+"/test_"+login+".rom")
                        defer f.Close()
			firmware := make([]byte,64*1024*1024)
                        _,_=f.Read(firmware)
			w.Write(firmware)
		case "getBMCFirmware":
                      login := tail[1:]
                        // We must retreive the username BIOS and return it as the response body
                        if ( OpenBMCCommand != nil ) {
                                unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
                                _ = <- OpenBMCBuildChannel
				OpenBMCCommand = nil
                        }
                        f, _ := os.Open(firmwaresPath+"/test_openbmc_"+login+".mtd")
                        defer f.Close()
                        firmware := make([]byte,32*1024*1024)
                        _,_=f.Read(firmware)
                        w.Write(firmware)
		case "buildbmcfirmware":
                        switch r.Method {
                                case http.MethodPut:
				        if ( OpenBMCCommand != nil ) {
                        		        unix.Kill(OpenBMCCommand.Process.Pid, unix.SIGINT)
						_ = <- OpenBMCBuildChannel
						OpenBMCCommand = nil
                        		}
					username := tail[1:]
                                        data := base.HTTPGetBody(r)
                                        keywords := strings.Fields(string(data))
                                        githubRepo := keywords[0]
                                        githubBranch := keywords[1]
                                        recipes := keywords[2]
					interactive := keywords[3]
                                        proxy := os.Getenv("PROXY")
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
					args = append (args, username)
                                        args = append (args, githubRepo)
                                        args = append (args, githubBranch)
                                        args = append (args, recipes)
					args = append (args, storageUri)
                                        args = append (args, storageTcpPort)
					args = append (args, interactive)

                                        args = append (args, proxy)
                                        OpenBMCCommand = exec.Command(startOpenBMCBuildBin, args...)
                                        OpenBMCCommand.SysProcAttr = &unix.SysProcAttr{
                                                Setsid: true,
                                        }
                                        err := OpenBMCCommand.Start()
					if ( err == nil ) {
	                                        go func() {
	                                                OpenBMCCommand.Wait()
							OpenBMCBuildChannel <- "done"
	                                        }()
						if ( interactive == "1" ) {
							// We must hang off after being sure that the console daemon is properly starter
		                                        conn, err := net.DialTimeout("tcp", "localhost:7682", 220*time.Millisecond)
		                                        max_loop := 5
		                                        for ( err != nil && max_loop > 0 ) {
		                                                conn, err = net.DialTimeout("tcp", "localhost:7682", 220*time.Millisecond)
								max_loop = max_loop - 1
		                                        }
		                                        if ( err != nil ) {
		                                        // Daemon has not started
		                                        // Let's report an error
		                                                w.Write([]byte("Error"))
		                                                return
		                                        } else {
		                                                conn.Close()
		                                        }
						}
					} else {
						OpenBMCCommand = nil
					}

				}
		case "buildbiosfirmware":
			switch r.Method {
                                case http.MethodPut:
					if ( LinuxBOOTCommand != nil ) {
                                                unix.Kill(LinuxBOOTCommand.Process.Pid, unix.SIGINT)
                                                _ = <- LinuxBOOTBuildChannel
						LinuxBOOTCommand = nil
                                        }
					username := tail[1:]
					data := base.HTTPGetBody(r)
					keywords := strings.Fields(string(data))
					githubRepo := keywords[0]
					githubBranch := keywords[1]
					board := keywords[2]
					interactive := keywords[3]
					proxy := os.Getenv("PROXY")
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
                                        args = append (args, username)
                                        args = append (args, githubRepo)
                                        args = append (args, githubBranch)
                                        args = append (args, board)
                                        args = append (args, storageUri)
                                        args = append (args, storageTcpPort)
					args = append (args, interactive)

                                        args = append (args, proxy)

                                        for i := 0 ; i < len(args) ; i++ {
                                                print(args[i]+"\n")
                                        }

                                        LinuxBOOTCommand = exec.Command(startLinuxbootBuildBin, args...)
                                        err := LinuxBOOTCommand.Start()
					if ( err == nil ) {	
	                                        go func() {
							LinuxBOOTCommand.Wait()
							LinuxBOOTBuildChannel <- "done"
						}()
						if ( interactive == "1" ) {
							// We must hang off after being sure that the console daemon is properly starter
		                                        conn, err := net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
		                                        max_loop := 5
		                                        for ( err != nil && max_loop > 0 ) {
		                                                conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
								max_loop = max_loop - 1
		                                        }
		                                        if ( err != nil ) {
			                                        // Daemon has not started
		                                        	// Let's report an error
		                                                w.Write([]byte("Error"))
		                                                return
		                                        } else {
		                                                conn.Close()
		                                        }
						}
					} else {
						LinuxBOOTCommand = nil
					}
				}
		default:
	}
}


func main() {
    print("=============================== \n")
    print("| Starting Compile backen     |\n")
    print("| Development version -       |\n")
    print("=============================== \n")

    dockerClient,_ = client.NewEnvClient()


    OpenBMCBuildChannel = make(chan string)
    LinuxBOOTBuildChannel = make(chan string)
    mux := http.NewServeMux()

    // Highest priority must be set to the signed request
    mux.HandleFunc("/",home)


    log.Fatal(http.ListenAndServe(compileTcpPort, mux))
}
