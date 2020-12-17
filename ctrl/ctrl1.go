package main

import (
	"net/http"
	"strings"
	"path"
	"fmt"
	"log"
	"os/exec"
	"net"
	"time"
	"os"
	"io"
	"base"
	"golang.org/x/sys/unix"
)

var binariesPath = os.Getenv("BINARIES_PATH")
var firmwaresPath = os.Getenv("FIRMWARES_PATH")
var distrosPath = os.Getenv("DISTROS_PATH")
var compileUri = os.Getenv("COMPILE_URI")
var compileTcpPort = os.Getenv("COMPILE_TCPPORT")
var storageUri = os.Getenv("STORAGE_URI")
var storageTcpPort = os.Getenv("STORAGE_TCPPORT")
var isEmulatorsPool = os.Getenv("IS_EMULATORS_POOL")
var em100Bios = os.Getenv("EM100BIOS")
var em100Bmc = os.Getenv("EM100BMC")
var bmcSerial = os.Getenv("BMC_SERIAL")

var OpenBMCEm100Command *exec.Cmd = nil
var bmcSerialConsoleCmd *exec.Cmd = nil
var RomEm100Command *exec.Cmd = nil

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
		case "getosinstallers":
			_,tail := ShiftPath( r.URL.Path)
			file :=  strings.Split(tail,"/")
			// file[1] does contain the name of the server which is needed
			// The file is seating into the storage server ... We have to transfer it
			// into a local storage (ideally a RAMFS to avoid potential storage impact
			// and accelerating transfer)
			// but all of this is performed within an external script as it needs to
			// be piped to a ttyd as to provide end user feedback
			fmt.Printf("Usb load received\n")
                        args := []string { distrosPath+"/"+file[1] }
                        cmd := exec.Command(binariesPath+"/load_usb", args...)
			cmd.SysProcAttr = &unix.SysProcAttr{
                                                Setsid: true,
                        }
                        cmd.Start()
                        done := make(chan error, 1)
                        go func() {
				done <- cmd.Wait()
			}()
		case "isEmulatorsPool":
			w.Write([]byte("{ \"isPool\":\""+isEmulatorsPool+"\" }"))
		case "resetEmulator":
			_,tail := ShiftPath( r.URL.Path)
			path :=  strings.Split(tail,"/")
                        emulator := path[1]
			if ( emulator == "bmc" ) {
				// We need to switch off the em100 associated to the BMC
				// This could be done by sending a kill signal to the associates ttyCommand if it does exist
				// and then reset the associated em100 through binariesPath/reset_em100 script
				if ( OpenBMCEm100Command != nil ) {
					unix.Kill(OpenBMCEm100Command.Process.Pid, unix.SIGTERM)
					OpenBMCEm100Command = nil
					var argsConsole []string
                                        argsConsole = append(argsConsole, "bmc")
                                        resetEm100Cmd := exec.Command(binariesPath+"/reset_em100", argsConsole...)
                                        resetEm100Cmd.Start()
                                        go func() {
                                                resetEm100Cmd.Wait()
                                        }()
				}
				if ( bmcSerialConsoleCmd != nil ) {
					unix.Kill(bmcSerialConsoleCmd.Process.Pid, unix.SIGTERM)
                                        bmcSerialConsoleCmd = nil
				}
				
                        } else {
                                if ( emulator == "rom" ) {
					if ( RomEm100Command != nil ) {
						unix.Kill(RomEm100Command.Process.Pid, unix.SIGTERM)
	                                        RomEm100Command = nil
						var argsConsole []string
	                                        argsConsole = append(argsConsole, "rom")
	                                        resetEm100Cmd := exec.Command(binariesPath+"/reset_em100", argsConsole...)
	                                        resetEm100Cmd.Start()
	                                        go func() {
	                                                resetEm100Cmd.Wait()
	                                        }()
					}
                                } else {
                                        w.Write([]byte(emulator))
                                }
                        }
		case "bmcfirmware":
                        switch r.Method {
                                case http.MethodPost:
					_,tail := ShiftPath( r.URL.Path)
                                        path :=  strings.Split(tail,"/")
                                        username := path[1]
                                        r.Body = http.MaxBytesReader(w, r.Body, 64<<20+4096)
                                        err := r.ParseMultipartForm(64<<20+4096)
                                        if ( err != nil ) {
                                                fmt.Printf("Error %s\n",err.Error())
                                        }
                                        file,handler,_ := r.FormFile("fichier")

                                        defer file.Close()
                                        f, err := os.OpenFile(firmwaresPath+"/_"+username+"_"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
                                        if err != nil {
                                               fmt.Println(err)
                                               return
                                        }
                                        defer f.Close()
                                        io.Copy(f, file)
					// we must forward the request to the relevant test server
		                        fmt.Printf("Ilo start received\n")

					var args []string
                                        args = append(args,"-p")
                                        args = append(args,"7681")
					args = append(args, "-s")
                                        args = append(args, "9")
                                        args = append(args,"-R")
                                        args = append(args,"unbuffer")
                                        args = append(args,binariesPath + "/em100")
                                        args = append(args,"-c")
                                        args = append(args,"MX25L25635E")
                                        args = append(args,"-x")
                                        args = append(args,em100Bmc)
                                        args = append(args,"-T")
                                        args = append(args,"-d")
                                        args = append(args, firmwaresPath+"/_"+username+"_"+handler.Filename)
                                        args = append(args,"-r")
                                        args = append(args,"-v")
                                        args = append(args,"-O")
                                        args = append(args,"0xFE0000000")
                                        args = append(args,"-p")
                                        args = append(args,"low")
                                        OpenBMCEm100Command = exec.Command(binariesPath+"/ttyd", args...)
		                        OpenBMCEm100Command.Start()

					// BMC console needs to be started also

					var argsConsole []string
					argsConsole = append(argsConsole, "-p")
					argsConsole = append(argsConsole, "7682")
					argsConsole = append(argsConsole, "-s")
					argsConsole = append(argsConsole, "9")
					argsConsole = append(argsConsole, "screen")
					argsConsole = append(argsConsole, bmcSerial)
					argsConsole = append(argsConsole, "115200")
					bmcSerialConsoleCmd = exec.Command(binariesPath+"/ttyd", argsConsole...)
					bmcSerialConsoleCmd.Start()

					go func() {
						bmcSerialConsoleCmd.Wait()
					}()

		                        done := make(chan error, 1)
		                        go func() {
		                            done <- OpenBMCEm100Command.Wait()
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
		case "biosfirmware":
			switch r.Method {
                                case http.MethodPost:
					_,tail := ShiftPath( r.URL.Path)
					path :=  strings.Split(tail,"/")
					username := path[1]
                                        r.Body = http.MaxBytesReader(w, r.Body, 64<<20+4096)
                                        err := r.ParseMultipartForm(64<<20+4096)
                                        if ( err != nil ) {
                                                fmt.Printf("Error %s\n",err.Error())
                                        }
                                        file,handler,_ := r.FormFile("fichier")

                                        defer file.Close()
                                        f, err := os.OpenFile(firmwaresPath+"/_"+username+"_"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
                                        if err != nil {
                                               fmt.Println(err)
                                               return
                                        }
                                        defer f.Close()
                                        io.Copy(f, file)
					// we must forward the request to the relevant test server
		                        fmt.Printf("System BIOS start received\n")
		                        var args []string
                        		args = append(args,"-p")
		                        args = append(args,"7683")
					args = append(args, "-s")
                                        args = append(args, "9")
		                        args = append(args,"-R")
		                        args = append(args,"unbuffer")
		                        args = append(args,binariesPath + "/em100")
		                        args = append(args,"-c")
               		         	args = append(args,"MX25L51245G")
                       		 	args = append(args,"-x")
		                        args = append(args,em100Bios)
		                        args = append(args,"-T")
		                        args = append(args,"-d")
		                        args = append(args, firmwaresPath+"/_"+username+"_"+handler.Filename)
		                        args = append(args,"-r")
		                        args = append(args,"-v")
		                        args = append(args,"-O")
		                        args = append(args,"0xFE0000000")
		                        args = append(args,"-p")
		                        args = append(args,"low")
		                        RomEm100Command := exec.Command(binariesPath+"/ttyd", args...)

		                        RomEm100Command.Start()
		                        done := make(chan error, 1)
		                        go func() {
		                            done <- RomEm100Command.Wait()
		                        }()
		                        conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
		                        max_loop := 5
		                        for ( err != nil && max_loop > 0 ) {
		                                conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
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
		case "loadfromstoragesmbios":
			// We must get the username from the request
			_, tail := ShiftPath( r.URL.Path)
                        login := tail[1:]
			// We have to retreive the BIOS from the compile server
			
                        _ = base.HTTPGetRequest("http://"+compileUri+compileTcpPort+"/cleanUp/rom")
			myfirmware := base.HTTPGetRequest("http://"+storageUri + storageTcpPort + "/user/"+login+"/getFirmware")
                        // f, err := os.Create("firmwares/linuxboot_"+login+".rom", os.O_WRONLY|os.O_CREATE, 0666)
                        f, err := os.Create(firmwaresPath+"/linuxboot_"+login+".rom")
			defer f.Close()
			f.Write([]byte(myfirmware))

			fmt.Printf("System BIOS start received\n")
                        var args []string
                        args = append(args,"-p")
                        args = append(args,"7683")
			args = append(args, "-s")
                        args = append(args, "9")
                        args = append(args,"-R")
                        args = append(args,"unbuffer")
                        args = append(args,binariesPath + "/em100")
                        args = append(args,"-c")
                        args = append(args,"MX25L51245G")
                        args = append(args,"-x")
                        args = append(args,em100Bios)
                        args = append(args,"-T")
                        args = append(args,"-d")
                        args = append(args, firmwaresPath+"/linuxboot_"+login+".rom")
                        args = append(args,"-r")
                        args = append(args,"-v")
                        args = append(args,"-O")
                        args = append(args,"0xFE0000000")
                        args = append(args,"-p")
                        args = append(args,"low")
                        RomEm100Command := exec.Command(binariesPath+"/ttyd", args...)
                        RomEm100Command.Start()
                        done := make(chan error, 1)
                        go func() {
       				done <- RomEm100Command.Wait()
                        }()
			// We need to wait that the process spawn before checking if it is up and running
			// total wait time can be up to 3s
			time.Sleep(2)
                        conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
                        max_loop := 5
                        for ( err != nil && max_loop > 0 ) {
				conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
			}
			if ( err != nil ) {
				// Daemon has not started
				// Let's report an error
					w.Write([]byte("Error"))
                                        return
                         } else {
				conn.Close()
                         }
		case "loadfromstoragebmc":
                        // We must get the username from the request
                        _, tail := ShiftPath( r.URL.Path)
                        login := tail[1:]
                        // We have to retreive the BIOS from the storage server

                        _ = base.HTTPGetRequest("http://"+compileUri+compileTcpPort+"/cleanUp/bmc")
                        myfirmware := base.HTTPGetRequest("http://"+storageUri + storageTcpPort + "/user/"+login+"/getBMCFirmware")
                        // f, err := os.Create("firmwares/openbmc_"+login+".rom", os.O_WRONLY|os.O_CREATE, 0666)
                        f, err := os.Create(firmwaresPath+"/openbmc_"+login+".rom")
                        defer f.Close()
                        f.Write([]byte(myfirmware))
                        fmt.Printf("BMC start received\n")

			var args []string
                        args = append(args,"-p")
                        args = append(args,"7681")
			args = append(args, "-s")
                        args = append(args, "9")
                        args = append(args,"-R")
                        args = append(args,"unbuffer")
                        args = append(args,binariesPath + "/em100")
                        args = append(args,"-c")
                        args = append(args,"MX25L25635E")
                        args = append(args,"-x")
                        args = append(args,em100Bmc)
                        args = append(args,"-T")
                        args = append(args,"-d")
                        args = append(args, firmwaresPath+"/openbmc_"+login+".rom")
                        args = append(args,"-r")
                        args = append(args,"-v")
                        args = append(args,"-O")
                        args = append(args,"0xFE0000000")
                        args = append(args,"-p")
                        args = append(args,"low")
                        OpenBMCEm100Command = exec.Command(binariesPath+"/ttyd", args...)
                        OpenBMCEm100Command.Start()

			// We need to start the console also

			var argsConsole []string
                        argsConsole = append(argsConsole, "-p")
                        argsConsole = append(argsConsole, "7682")
			argsConsole = append(argsConsole, "-s")
                        argsConsole = append(argsConsole, "9")
                        argsConsole = append(argsConsole, "screen")
                        argsConsole = append(argsConsole, bmcSerial)
                        argsConsole = append(argsConsole, "115200")
                        bmcSerialConsoleCmd = exec.Command(binariesPath+"/ttyd", argsConsole...)
                        bmcSerialConsoleCmd.Start()

                        go func() {
				bmcSerialConsoleCmd.Wait()
                        }()

                        done := make(chan error, 1)
                        go func() {
                               done <- OpenBMCEm100Command.Wait()
                        }()
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
		case "startbmc":
			fmt.Printf("BMC start received\n")
			var args []string
                        args = append(args,"-p")
                        args = append(args,"7681")
			args = append(args, "-s")
                        args = append(args, "9")
                        args = append(args,"-R")
                        args = append(args,"unbuffer")
                        args = append(args,binariesPath + "/em100")
                        args = append(args,"-c")
                        args = append(args,"MX25L25635E")
                        args = append(args,"-x")
                        args = append(args,em100Bmc)
                        args = append(args,"-T")
                        args = append(args,"-d")
                        args = append(args, firmwaresPath+"/ilo_dl360_OpenBMC.rom")
                        args = append(args,"-r")
                        args = append(args,"-v")
                        args = append(args,"-O")
                        args = append(args,"0xFE0000000")
                        args = append(args,"-p")
                        args = append(args,"low")
                        OpenBMCEm100Command = exec.Command(binariesPath+"/ttyd", args...)
                        OpenBMCEm100Command.Start()

			// we need to start also the console

			var argsConsole []string
                        argsConsole = append(argsConsole, "-p")
                        argsConsole = append(argsConsole, "7682")
			argsConsole = append(argsConsole, "-s")
                        argsConsole = append(argsConsole, "9")
                        argsConsole = append(argsConsole, "screen")
                        argsConsole = append(argsConsole, bmcSerial)
                        argsConsole = append(argsConsole, "115200")
                        bmcSerialConsoleCmd = exec.Command(binariesPath+"/ttyd", argsConsole...)
                        bmcSerialConsoleCmd.Start()

                        go func() {
                                bmcSerialConsoleCmd.Wait()
                        }()

			done := make(chan error, 1)
                        go func() {
                            done <- OpenBMCEm100Command.Wait()
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
		case "startsmbios":
			// we must forward the request to the relevant test server
                        fmt.Printf("System BIOS start received\n")
                        var args []string
                        args = append(args,"-p")
                        args = append(args,"7683")
			args = append(args, "-s")
                        args = append(args, "9")
                        args = append(args,"-R")
                        args = append(args,"unbuffer")
                        args = append(args,binariesPath + "/em100")
                        args = append(args,"-c")
                        args = append(args,"MX25L51245G")
                        args = append(args,"-x")
                        args = append(args,em100Bios)
                        args = append(args,"-T")
                        args = append(args,"-d")
                        args = append(args, firmwaresPath+"/SBIOS_OpenBMC.rom")
                        args = append(args,"-r")
                        args = append(args,"-v")
                        args = append(args,"-O")
                        args = append(args,"0xFE0000000")
                        args = append(args,"-p")
                        args = append(args,"low")
                        RomEm100Command = exec.Command(binariesPath+"/ttyd", args...)
                        RomEm100Command.Start()
                        done := make(chan error, 1)
                        go func() {
                            done <- RomEm100Command.Wait()
                        }()
			conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
                        max_loop := 5
                        for ( err != nil && max_loop > 0 ) {
                                conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
                        }
                        if ( err != nil ) {
                                // Daemon has not started
                                // Let's report an error
                                w.Write([]byte("Error"))
                                return
                        } else {
                                conn.Close()
                        }
		case "poweron":
			fmt.Printf("start power\n")
                        args := []string { "on" }
                        cmd := exec.Command(binariesPath+"/iPDUpower", args...)
                        cmd.Start()
                        done := make(chan error, 1)
                        go func() {
                            done <- cmd.Wait()
                        }()
		case "bmcup":
			
		case "poweroff":
			// We need to cleanup the em100
			// We also need to clean up the screen command
			// and free the USB->Serial
			fmt.Printf("stop power\n")
                        args := []string { "off" }
                        cmd := exec.Command(binariesPath+"/iPDUpower", args...)
                        cmd.Start()
                        cmd.Wait()
			args = []string { "" }
                        cmd = exec.Command(binariesPath+"/cleanUP", args...)
                        cmd.Start()
                        cmd.Wait()
		default:
	}
}

func main() {
    print("=============================== \n")
    print("| Starting frontend           |\n")
    print("| Development version -       |\n")
    print("=============================== \n")

    var ctrlTcpPort = os.Getenv("CTRL_TCPPORT")
    mux := http.NewServeMux()

    // Highest priority must be set to the signed request
    mux.HandleFunc("/",home)


    log.Fatal(http.ListenAndServe(ctrlTcpPort, mux))
}
