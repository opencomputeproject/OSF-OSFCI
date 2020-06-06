package main

import (
	"net/http"
	"strings"
	"path"
	"log"
        "os"
	"os/exec"
	"base"
	"golang.org/x/sys/unix"
)

var compileTcpPort = os.Getenv("COMPILE_TCPPORT")
var startLinuxbootBuildBin = os.Getenv("LINUXBOOT_BUILD")
var binariesPath = os.Getenv("BINARIES_PATH")
var firmwaresPath = os.Getenv("FIRMWARES_PATH")
var ttydCommand *exec.Cmd = nil
var dockerCommand *exec.Cmd = nil

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
			if ( ttydCommand != nil ) {
                                unix.Kill(ttydCommand.Process.Pid, unix.SIGINT)
                        }
                        if ( dockerCommand != nil ) {
                                unix.Kill(-dockerCommand.Process.Pid, unix.SIGKILL)
                                unix.Kill(dockerCommand.Process.Pid, unix.SIGKILL)
                        }
		case "getFirmware":
			login := tail[1:]
			// We must retreive the username BIOS and return it as the response body
			if ( ttydCommand != nil ) {
				unix.Kill(ttydCommand.Process.Pid, unix.SIGINT)
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
		case "buildilofirmware":
                        switch r.Method {
                                case http.MethodPut:
				}
		case "buildbiosfirmware":
			switch r.Method {
                                case http.MethodPut:
					username := tail[1:]
					data := base.HTTPGetBody(r)
					keywords := strings.Fields(string(data))
					githubRepo := keywords[0]
					githubBranch := keywords[1]
					board := keywords[2]
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
                                        ttydCommand = exec.Command(binariesPath + "/ttyd", argsTtyd...)
					ttydCommand.SysProcAttr = &unix.SysProcAttr{
                                                Setsid: true,
                                        }
                                        ttydCommand.Start()
                                        go func() {
						ttydCommand.Wait()
						// This command is respinning itself
					}()

                                        var args []string
                                        args = append (args, username)
                                        args = append (args, githubRepo)
                                        args = append (args, githubBranch)
                                        args = append (args, board)
                                        for i := 0 ; i < len(args) ; i++ {
                                                print(args[i]+"\n")
                                        }

                                        dockerCommand = exec.Command(startLinuxbootBuildBin, args...)
                                        dockerCommand.Start()
                                        go func() {
						dockerCommand.Wait()
					}()

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


    log.Fatal(http.ListenAndServe(compileTcpPort, mux))
}
