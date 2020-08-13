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
	"golang.org/x/sys/unix"
)

var compileTcpPort = os.Getenv("COMPILE_TCPPORT")
var startLinuxbootBuildBin = os.Getenv("LINUXBOOT_BUILD")
var startOpenBMCBuildBin = os.Getenv("OPENBMC_BUILD")
var binariesPath = os.Getenv("BINARIES_PATH")
var firmwaresPath = os.Getenv("FIRMWARES_PATH")
var ttydCommandlinuxboot *exec.Cmd = nil
var ttydCommandopenbmc *exec.Cmd = nil
var dockerCommand *exec.Cmd = nil
var OpenBMCBuildChannel chan string

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


func home(w http.ResponseWriter, r *http.Request) {
	head,tail := ShiftPath( r.URL.Path)
	switch ( head ) {
		case "cleanUp":
			if ( ttydCommandlinuxboot != nil ) {
                                unix.Kill(ttydCommandlinuxboot.Process.Pid, unix.SIGINT)
                        }
			if ( ttydCommandopenbmc != nil ) {
                                unix.Kill(ttydCommandopenbmc.Process.Pid, unix.SIGINT)
                        }
                        if ( dockerCommand != nil ) {
                                unix.Kill(-dockerCommand.Process.Pid, unix.SIGKILL)
                                unix.Kill(dockerCommand.Process.Pid, unix.SIGKILL)
                        }
		case "getFirmware":
			login := tail[1:]
			// We must retreive the username BIOS and return it as the response body
			if ( ttydCommandlinuxboot != nil ) {
				unix.Kill(ttydCommandlinuxboot.Process.Pid, unix.SIGINT)
                        }
                        if ( dockerCommand != nil ) {
				unix.Kill(-dockerCommand.Process.Pid, unix.SIGKILL)
                                unix.Kill(dockerCommand.Process.Pid, unix.SIGKILL)
                        }
			f, _ := os.Open(firmwaresPath+"/test_"+login+".rom")
                        defer f.Close()
			firmware := make([]byte,64*1024*1024)
                        _,_=f.Read(firmware)
			w.Write(firmware)
		case "getBMCFirmware":
                      login := tail[1:]
                        // We must retreive the username BIOS and return it as the response body
                        if ( ttydCommandopenbmc != nil ) {
                                unix.Kill(ttydCommandopenbmc.Process.Pid, unix.SIGINT)
                        }
                        f, _ := os.Open(firmwaresPath+"/test_openbmc_"+login+".mtd")
                        defer f.Close()
                        firmware := make([]byte,32*1024*1024)
                        _,_=f.Read(firmware)
                        w.Write(firmware)
		case "buildbmcfirmware":
                        switch r.Method {
                                case http.MethodPut:
				        if ( ttydCommandopenbmc != nil ) {
                        		        unix.Kill(ttydCommandopenbmc.Process.Pid, unix.SIGINT)
						_ = <- OpenBMCBuildChannel
                        		}
					username := tail[1:]
                                        data := base.HTTPGetBody(r)
                                        keywords := strings.Fields(string(data))
                                        githubRepo := keywords[0]
                                        githubBranch := keywords[1]
                                        recipes := keywords[2]
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
                                        args = append (args, proxy)
                                        ttydCommandopenbmc = exec.Command(startOpenBMCBuildBin, args...)
                                        ttydCommandopenbmc.SysProcAttr = &unix.SysProcAttr{
                                                Setsid: true,
                                        }
                                        ttydCommandopenbmc.Start()
                                        go func() {
                                                ttydCommandopenbmc.Wait()
						OpenBMCBuildChannel <- "done"
                                        }()
					// We must hang off after being sure that the console daemon is properly starter
                                        conn, err := net.DialTimeout("tcp", "localhost:7682", 220*time.Millisecond)
                                        max_loop := 5
                                        for ( err != nil && max_loop > 0 ) {
                                                conn, err = net.DialTimeout("tcp", "localhost:7682", 220*time.Millisecond)
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
		case "buildbiosfirmware":
			switch r.Method {
                                case http.MethodPut:
					if ( ttydCommandlinuxboot != nil ) {
                        		        unix.Kill(ttydCommandlinuxboot.Process.Pid, unix.SIGINT)
                        		}
					username := tail[1:]
					data := base.HTTPGetBody(r)
					keywords := strings.Fields(string(data))
					githubRepo := keywords[0]
					githubBranch := keywords[1]
					board := keywords[2]
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

                                        var argsTtyd []string
                                        argsTtyd = append (argsTtyd,"-p")
                                        argsTtyd = append (argsTtyd,"7681")
                                        argsTtyd = append (argsTtyd,"-s")
                                        argsTtyd = append (argsTtyd,"9")
                                        argsTtyd = append (argsTtyd,binariesPath+"/readBiosFifo")
                                        ttydCommandlinuxboot = exec.Command(binariesPath + "/ttyd", argsTtyd...)
					ttydCommandlinuxboot.SysProcAttr = &unix.SysProcAttr{
                                                Setsid: true,
                                        }
                                        ttydCommandlinuxboot.Start()
                                        go func() {
						ttydCommandlinuxboot.Wait()
						// This command is respinning itself
					}()

                                        var args []string
                                        args = append (args, username)
                                        args = append (args, githubRepo)
                                        args = append (args, githubBranch)
                                        args = append (args, board)
                                        args = append (args, proxy)
                                        for i := 0 ; i < len(args) ; i++ {
                                                print(args[i]+"\n")
                                        }

                                        dockerCommand = exec.Command(startLinuxbootBuildBin, args...)
                                        dockerCommand.Start()
                                        go func() {
						dockerCommand.Wait()
					}()
					// We must hang off after being sure that the console daemon is properly starter
                                        conn, err := net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
                                        max_loop := 5
                                        for ( err != nil && max_loop > 0 ) {
                                                conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
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
		default:
	}
}


func main() {
    print("=============================== \n")
    print("| Starting frontend           |\n")
    print("| Development version -       |\n")
    print("=============================== \n")

    OpenBMCBuildChannel = make(chan string)
    mux := http.NewServeMux()

    // Highest priority must be set to the signed request
    mux.HandleFunc("/",home)


    log.Fatal(http.ListenAndServe(compileTcpPort, mux))
}
