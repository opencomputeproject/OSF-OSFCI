// OSFCI Controller module

package main

import (
	"base/base"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

var binariesPath string
var firmwaresPath string
var distrosPath string
var compileURI string
var compileTCPPort string
var storageURI string
var storageTCPPort string
var isEmulatorsPool string
var em100Bios string
var em100Bmc string
var bmcSerial string
var originalBmc string
var originalBios string
var bmcrecipe string
var biosrecipe string
var testPath string
var testLog string
var contestServer string
var solLogPath string
var bmcChip string
var biosChip string

// OpenBMCEm100Command string
var OpenBMCEm100Command *exec.Cmd = nil
var bmcSerialConsoleCmd *exec.Cmd = nil

// RomEm100Command string
var RomEm100Command *exec.Cmd = nil
var romSerialConsoleCmd *exec.Cmd = nil

// Test Console ttyd
var contestStartCmd *exec.Cmd = nil

// TestLists holds test-case(s) info
type TestLists struct {
	Name string
	Path string
}

// Initialize controller1 config
func initCtrlconfig() error {
	viper.SetConfigName("ctrl1conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/production/config/")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	binariesPath = viper.GetString("BINARIES_PATH")
	firmwaresPath = viper.GetString("FIRMWARES_PATH")
	distrosPath = viper.GetString("DISTROS_PATH")
	compileURI = viper.GetString("COMPILE_URI")
	compileTCPPort = viper.GetString("COMPILE_TCPPORT")
	storageURI = viper.GetString("STORAGE_URI")
	storageTCPPort = viper.GetString("STORAGE_TCPPORT")
	isEmulatorsPool = viper.GetString("IS_EMULATORS_POOL")
	em100Bios = viper.GetString("EM100BIOS")
	em100Bmc = viper.GetString("EM100BMC")
	bmcSerial = viper.GetString("BMC_SERIAL")
	originalBmc = viper.GetString("ORIGINAL_BMC")
	originalBios = viper.GetString("ORIGINAL_BIOS")
	bmcrecipe = viper.GetString("BMC_RECIPE")
	biosrecipe = viper.GetString("BIOS_RECIPE")
	testPath = viper.GetString("TEST_PATH")
	testLog = viper.GetString("TEST_LOG")
	contestServer = viper.GetString("CONTEST_SERVER")
	solLogPath = viper.GetString("SOL_LOG")
	bmcChip = viper.GetString("BMC_CHIP")
	biosChip = viper.GetString("BIOS_CHIP")

	return nil
}

// ShiftPath cleans up path
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func home(w http.ResponseWriter, r *http.Request) {
	head, _ := ShiftPath(r.URL.Path)
	switch head {
	case "get_os_installers":
		_, tail := ShiftPath(r.URL.Path)
		file := strings.Split(tail, "/")
		// file[1] does contain the name of the server which is needed
		// The file is seating into the storage server ... We have to transfer it
		// into a local storage (ideally a RAMFS to avoid potential storage impact
		// and accelerating transfer)
		// but all of this is performed within an external script as it needs to
		// be piped to a ttyd as to provide end user feedback
		fmt.Printf("Usb load received\n")
		args := []string{distrosPath + "/" + file[1]}
		cmd := exec.Command(binariesPath+"/load_usb", args...)
		cmd.SysProcAttr = &unix.SysProcAttr{
			Setsid: true,
		}
		cmd.Start()
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
	case "is_emulators_pool":
		w.Write([]byte("{ \"isPool\":\"" + isEmulatorsPool + "\" }"))
	case "reset_emulator":
		_, tail := ShiftPath(r.URL.Path)
		path := strings.Split(tail, "/")
		emulator := path[1]
		if emulator == "bmc" {
			// We need to switch off the em100 associated to the BMC
			// This could be done by sending a kill signal to the associates ttyCommand if it does exist
			// and then reset the associated em100 through binariesPath/reset_em100 script
			if OpenBMCEm100Command != nil {
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
			if bmcSerialConsoleCmd != nil {
				unix.Kill(bmcSerialConsoleCmd.Process.Pid, unix.SIGTERM)
				bmcSerialConsoleCmd = nil
			}

		} else {
			if emulator == "rom" {
				if RomEm100Command != nil {
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
	case "bmc_firmware":
		switch r.Method {
		case http.MethodPost:
			_, tail := ShiftPath(r.URL.Path)
			path := strings.Split(tail, "/")
			username := path[1]
			r.Body = http.MaxBytesReader(w, r.Body, 64<<20+4096)
			err := r.ParseMultipartForm(64<<20 + 4096)
			if err != nil {
				fmt.Printf("Error %s\n", err.Error())
			}
			file, handler, _ := r.FormFile("fichier")

			defer file.Close()
			f, err := os.OpenFile(firmwaresPath+"/_"+username+"_"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			io.Copy(f, file)

			// Truncate the file, if already exists
			if _, fileerr := os.Stat(solLogPath + "openbmc_sol_" + username + ".log"); fileerr == nil {
				os.Truncate(solLogPath+"openbmc_sol_"+username+".log", 0)
			}
			// we must forward the request to the relevant test server
			fmt.Printf("Ilo start received\n")

			var args []string
			args = append(args, "-p")
			args = append(args, "7681")
			args = append(args, "-s")
			args = append(args, "9")
			args = append(args, "-R")
			args = append(args, "unbuffer")
			args = append(args, binariesPath+"/em100")
			args = append(args, "-c")
			args = append(args, bmcChip)
			args = append(args, "-x")
			args = append(args, em100Bmc)
			args = append(args, "-T")
			args = append(args, "-d")
			args = append(args, firmwaresPath+"/_"+username+"_"+handler.Filename)
			args = append(args, "-r")
			args = append(args, "-v")
			args = append(args, "-O")
			args = append(args, "0xFE0000000")
			args = append(args, "-p")
			args = append(args, "low")
			OpenBMCEm100Command = exec.Command(binariesPath+"/ttyd", args...)
			OpenBMCEm100Command.Start()

			// BMC console needs to be started also

			var argsConsole []string
			argsConsole = append(argsConsole, "-p")
			argsConsole = append(argsConsole, "7682")
			argsConsole = append(argsConsole, "-s")
			argsConsole = append(argsConsole, "9")
			argsConsole = append(argsConsole, "screen")
			argsConsole = append(argsConsole, "-L")
			argsConsole = append(argsConsole, "-Logfile")
			argsConsole = append(argsConsole, solLogPath+"openbmc_sol_"+username+".log")
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
			maxLoop := 5
			for err != nil && maxLoop > 0 {
				conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
			}
			if err != nil {
				// Daemon has not started
				// Let's report an error
				w.Write([]byte("Error"))
				return
			}
			conn.Close()
		}

	case "bios_firmware":
		switch r.Method {
		case http.MethodPost:
			_, tail := ShiftPath(r.URL.Path)
			path := strings.Split(tail, "/")
			username := path[1]
			r.Body = http.MaxBytesReader(w, r.Body, 64<<20+4096)
			err := r.ParseMultipartForm(64<<20 + 4096)
			if err != nil {
				fmt.Printf("Error %s\n", err.Error())
			}
			file, handler, _ := r.FormFile("fichier")

			defer file.Close()
			f, err := os.OpenFile(firmwaresPath+"/_"+username+"_"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			io.Copy(f, file)
			// we must forward the request to the relevant test server
			fmt.Printf("System BIOS start received\n")
			var args []string
			args = append(args, "-p")
			args = append(args, "7683")
			args = append(args, "-s")
			args = append(args, "9")
			args = append(args, "-R")
			args = append(args, "unbuffer")
			args = append(args, binariesPath+"/em100")
			args = append(args, "-c")
			args = append(args, biosChip)
			args = append(args, "-x")
			args = append(args, em100Bios)
			args = append(args, "-T")
			args = append(args, "-d")
			args = append(args, firmwaresPath+"/_"+username+"_"+handler.Filename)
			args = append(args, "-r")
			args = append(args, "-v")
			args = append(args, "-O")
			args = append(args, "0xFE0000000")
			args = append(args, "-p")
			args = append(args, "low")
			RomEm100Command := exec.Command(binariesPath+"/ttyd", args...)

			RomEm100Command.Start()
			done := make(chan error, 1)
			go func() {
				done <- RomEm100Command.Wait()
			}()
			conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
			maxLoop := 5
			for err != nil && maxLoop > 0 {
				conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
			}
			if err != nil {
				// Daemon has not started
				// Let's report an error
				w.Write([]byte("Error"))
				return
			}
			conn.Close()
		}

	case "load_from_storage_smbios":
		// We must get the username from the request
		_, tail := ShiftPath(r.URL.Path)
		login := tail[1:]
		// We have to retrieve the BIOS from the compile server

		_ = base.HTTPGetRequest("http://" + compileURI + compileTCPPort + "/clean_up/rom")
		myfirmware := base.HTTPGetRequest("http://" + storageURI + storageTCPPort + "/user/" + login + "/get_firmware/" + biosrecipe + "/")
		// f, err := os.Create("firmwares/linuxboot_"+login+".rom", os.O_WRONLY|os.O_CREATE, 0666)
		f, err := os.Create(firmwaresPath + "/linuxboot_" + login + ".rom")
		defer f.Close()
		f.Write([]byte(myfirmware))

		fmt.Printf("System BIOS start received\n")
		var args []string
		args = append(args, "-p")
		args = append(args, "7683")
		args = append(args, "-s")
		args = append(args, "9")
		args = append(args, "-R")
		args = append(args, "unbuffer")
		args = append(args, binariesPath+"/em100")
		args = append(args, "-c")
		args = append(args, biosChip)
		args = append(args, "-x")
		args = append(args, em100Bios)
		args = append(args, "-T")
		args = append(args, "-d")
		args = append(args, firmwaresPath+"/linuxboot_"+login+".rom")
		args = append(args, "-r")
		args = append(args, "-v")
		args = append(args, "-O")
		args = append(args, "0xFE0000000")
		args = append(args, "-p")
		args = append(args, "low")
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
		maxLoop := 5
		for err != nil && maxLoop > 0 {
			conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
		}
		if err != nil {
			// Daemon has not started
			// Let's report an error
			w.Write([]byte("Error"))
			return
		}
		conn.Close()

	case "load_from_storage_bmc":
		// We must get the username from the request
		_, tail := ShiftPath(r.URL.Path)
		login := tail[1:]
		// We have to retrieve the BIOS from the storage server

		_ = base.HTTPGetRequest("http://" + compileURI + compileTCPPort + "/clean_up/bmc")
		myfirmware := base.HTTPGetRequest("http://" + storageURI + storageTCPPort + "/user/" + login + "/get_bmc_firmware/" + bmcrecipe + "/")
		// f, err := os.Create("firmwares/openbmc_"+login+".rom", os.O_WRONLY|os.O_CREATE, 0666)
		f, err := os.Create(firmwaresPath + "/openbmc_" + login + ".rom")
		defer f.Close()
		f.Write([]byte(myfirmware))
		fmt.Printf("BMC start received\n")
		// Truncate the file, if already exists
		if _, fileerr := os.Stat(solLogPath + "openbmc_sol_" + login + ".log"); fileerr == nil {
			os.Truncate(solLogPath+"openbmc_sol_"+login+".log", 0)
		}

		var args []string
		args = append(args, "-p")
		args = append(args, "7681")
		args = append(args, "-s")
		args = append(args, "9")
		args = append(args, "-R")
		args = append(args, "unbuffer")
		args = append(args, binariesPath+"/em100")
		args = append(args, "-c")
		args = append(args, bmcChip)
		args = append(args, "-x")
		args = append(args, em100Bmc)
		args = append(args, "-T")
		args = append(args, "-d")
		args = append(args, firmwaresPath+"/openbmc_"+login+".rom")
		args = append(args, "-r")
		args = append(args, "-v")
		args = append(args, "-O")
		args = append(args, "0xFE0000000")
		args = append(args, "-p")
		args = append(args, "low")
		OpenBMCEm100Command = exec.Command(binariesPath+"/ttyd", args...)
		OpenBMCEm100Command.Start()

		// We need to start the console also

		var argsConsole []string
		argsConsole = append(argsConsole, "-p")
		argsConsole = append(argsConsole, "7682")
		argsConsole = append(argsConsole, "-s")
		argsConsole = append(argsConsole, "9")
		argsConsole = append(argsConsole, "screen")
		argsConsole = append(argsConsole, "-L")
		argsConsole = append(argsConsole, "-Logfile")
		argsConsole = append(argsConsole, solLogPath+"openbmc_sol_"+login+".log")
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
		maxLoop := 5
		for err != nil && maxLoop > 0 {
			conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
		}
		if err != nil {
			// Daemon has not started
			// Let's report an error
			w.Write([]byte("Error"))
			return
		}
		conn.Close()

	case "start_bmc":
		// We must get the username from the request
		_, tail := ShiftPath(r.URL.Path)
		username := tail[1:]
		fmt.Printf("BMC start received\n")
		// Truncate the file, if already exists
		if _, fileerr := os.Stat(solLogPath + "openbmc_sol_" + username + ".log"); fileerr == nil {
			os.Truncate(solLogPath+"openbmc_sol_"+username+".log", 0)
		}
		var args []string
		args = append(args, "-p")
		args = append(args, "7681")
		args = append(args, "-s")
		args = append(args, "9")
		args = append(args, "-R")
		args = append(args, "unbuffer")
		args = append(args, binariesPath+"/em100")
		args = append(args, "-c")
		args = append(args, bmcChip)
		args = append(args, "-x")
		args = append(args, em100Bmc)
		args = append(args, "-T")
		args = append(args, "-d")
		args = append(args, firmwaresPath+"/"+originalBmc)
		args = append(args, "-r")
		args = append(args, "-v")
		args = append(args, "-O")
		args = append(args, "0xFE0000000")
		args = append(args, "-p")
		args = append(args, "low")
		OpenBMCEm100Command = exec.Command(binariesPath+"/ttyd", args...)
		OpenBMCEm100Command.Start()

		// we need to start also the console

		var argsConsole []string
		argsConsole = append(argsConsole, "-p")
		argsConsole = append(argsConsole, "7682")
		argsConsole = append(argsConsole, "-s")
		argsConsole = append(argsConsole, "9")
		argsConsole = append(argsConsole, "screen")
		argsConsole = append(argsConsole, "-L")
		argsConsole = append(argsConsole, "-Logfile")
		argsConsole = append(argsConsole, solLogPath+"openbmc_sol_"+username+".log")
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
		maxLoop := 5
		for err != nil && maxLoop > 0 {
			conn, err = net.DialTimeout("tcp", "localhost:7681", 220*time.Millisecond)
		}
		if err != nil {
			// Daemon has not started
			// Let's report an error
			w.Write([]byte("Error"))
			return
		}
		conn.Close()

	case "rom_sol_log":
		base.Zlog.Infof("ROM sol log start received")
		// Truncate the file, if already exists
		if _, fileerr := os.Stat(solLogPath + "bios_sol.log"); fileerr == nil {
			os.Truncate(solLogPath+"bios_sol.log", 0)
		}

		// we need to start also the console
		sutIP := r.FormValue("bmcip")
		base.Zlog.Infof("SUTIP:", sutIP)

		var argsConsole []string
		argsConsole = append(argsConsole, sutIP)
		romSerialConsoleCmd = exec.Command(binariesPath+"/bioslog", argsConsole...)
		romSerialConsoleCmd.Start()
		go func() {
			romSerialConsoleCmd.Wait()
		}()
	case "start_smbios":
		// we must forward the request to the relevant test server
		fmt.Printf("System BIOS start received\n")
		var args []string
		args = append(args, "-p")
		args = append(args, "7683")
		args = append(args, "-s")
		args = append(args, "9")
		args = append(args, "-R")
		args = append(args, "unbuffer")
		args = append(args, binariesPath+"/em100")
		args = append(args, "-c")
		args = append(args, biosChip)
		args = append(args, "-x")
		args = append(args, em100Bios)
		args = append(args, "-T")
		args = append(args, "-d")
		args = append(args, firmwaresPath+"/"+originalBios)
		args = append(args, "-r")
		args = append(args, "-v")
		args = append(args, "-O")
		args = append(args, "0xFE0000000")
		args = append(args, "-p")
		args = append(args, "low")
		RomEm100Command = exec.Command(binariesPath+"/ttyd", args...)
		RomEm100Command.Start()
		done := make(chan error, 1)
		go func() {
			done <- RomEm100Command.Wait()
		}()
		conn, err := net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
		maxLoop := 5
		for err != nil && maxLoop > 0 {
			conn, err = net.DialTimeout("tcp", "localhost:7683", 220*time.Millisecond)
		}
		if err != nil {
			// Daemon has not started
			// Let's report an error
			w.Write([]byte("Error"))
			return
		}
		conn.Close()

	case "power_on":
		fmt.Printf("start power\n")
		args := []string{"on"}
		cmd := exec.Command(binariesPath+"/iPDUpower", args...)
		cmd.Start()
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

	case "bmc_up":

	case "power_off":
		// We need to cleanup the em100
		// We also need to clean up the screen command
		// and free the USB->Serial
		fmt.Printf("stop power\n")
		args := []string{"off"}
		cmd := exec.Command(binariesPath+"/iPDUpower", args...)
		cmd.Start()
		cmd.Wait()
		args = []string{""}
		cmd = exec.Command(binariesPath+"/cleanUP", args...)
		cmd.Start()
		cmd.Wait()
	case "test_list":
		testlists, err := getStandardTests()
		if err != nil {
			base.Zlog.Infof(err.Error())
			return
		}
		returnData, err := json.Marshal(testlists)
		if err != nil {
			base.Zlog.Infof(err.Error())
		}
		w.Write([]byte(returnData))
	case "test_start":
		base.Zlog.Infof("Starting OpenBMC Testing")
		_, tail := ShiftPath(r.URL.Path)
		path := strings.Split(tail, "/")
		username := path[1]
		base.Zlog.Infof("Username:%s", username)
		input := r.FormValue("testlist")
		var tests []string
		json.Unmarshal([]byte(input), &tests)
		if len(tests) < 1 {
			base.Zlog.Infof("Empty Tests")
			w.Write([]byte("Empty test lists"))
			return
		}
		testlist := strings.Join(tests, ",")
		base.Zlog.Infof("Testlist:", testlist)

		if contestStartCmd != nil {
			unix.Kill(contestStartCmd.Process.Pid, unix.SIGTERM)
			contestStartCmd = nil
		}
		var args []string
		args = append(args, "-p")
		args = append(args, "8081")
		args = append(args, "-s")
		args = append(args, "9")
		args = append(args, "-t")
		args = append(args, "disableReconnect=true")
		args = append(args, binariesPath+"/contestcli")
		args = append(args, "-user="+username)
		args = append(args, "-tests="+testlist)
		args = append(args, "-log="+testLog)
		args = append(args, "-addr="+contestServer)
		contestStartCmd = exec.Command(binariesPath+"/ttyd", args...)
		err := contestStartCmd.Start()
		if err != nil {
			base.Zlog.Infof(err.Error())
		}
		go func() {
			contestStartCmd.Wait()
		}()
		conn, err := net.DialTimeout("tcp", "localhost:8081", 220*time.Millisecond)
		maxLoop := 5
		for err != nil && maxLoop > 0 {
			conn, err = net.DialTimeout("tcp", "localhost:8081", 220*time.Millisecond)
		}
		conn.Close()
	case "test_logs":
		_, tail := ShiftPath(r.URL.Path)
		path := strings.Split(tail, "/")
		username := path[1]
		base.Zlog.Infof("Fetching the Test logs for user:%s", username)
		command := "ls -tp " + testLog + "/contest_" + username + "*.zip | head -1"
		base.Zlog.Infof("Executing command: %s", command)
		cmd := exec.Command("/bin/sh", "-c", command)
		out, err := cmd.Output()
		if err != nil {
			base.Zlog.Infof(err.Error())
			return
		}
		logfile := strings.TrimSpace(string(out[:]))
		base.Zlog.Infof(logfile)
		content, err := ioutil.ReadFile(logfile)
		if err != nil {
			base.Zlog.Infof(err.Error())
			return
		}
		w.Header().Add("Content-Length", strconv.Itoa(len(content)))
		w.Write(content)
	case "get_bmc_logs":
		_, tail := ShiftPath(r.URL.Path)
		path := strings.Split(tail, "/")
		username := path[1]
		base.Zlog.Infof("Fetching the SOL BMC logs for user:%s", username)
		logfile := solLogPath + "openbmc_sol_" + username + ".log"
		base.Zlog.Infof(logfile)
		content, err := ioutil.ReadFile(logfile)
		if err != nil {
			base.Zlog.Infof(err.Error())
			return
		}
		w.Header().Add("Content-Length", strconv.Itoa(len(content)))
		w.Write(content)
	case "get_bios_logs":
		base.Zlog.Infof("Fetching the BIOS logs")
		logfile := solLogPath + "bios_sol.log"
		base.Zlog.Infof(logfile)
		content, err := ioutil.ReadFile(logfile)
		if err != nil {
			base.Zlog.Infof(err.Error())
			return
		}
		w.Header().Add("Content-Length", strconv.Itoa(len(content)))
		w.Write(content)
	default:
	}
}

func getStandardTests() ([]*TestLists, error) {
	var testlists []*TestLists
	files, err := ioutil.ReadDir(testPath)
	if err != nil {
		base.Zlog.Infof(err.Error())
		return testlists, err
	}
	for _, file := range files {
		if file.IsDir() == false {
			testpath := path.Join(testPath, file.Name())
			testname, _ := getTestName(testpath)
			testlists = append(testlists, &TestLists{
				Name: testname,
				Path: testpath,
			})
		}
	}
	return testlists, nil
}

func getTestName(testpath string) (string, error) {
	jsoncontent, err := ioutil.ReadFile(testpath)
	if err != nil {
		base.Zlog.Infof(err.Error())
		return "", err
	}
	var jsontest map[string]interface{}
	json.Unmarshal([]byte(jsoncontent), &jsontest)
	if _, ok := jsontest["JobName"]; ok {
		return jsontest["JobName"].(string), nil
	}
	return path.Base(testpath), nil
}

// Default Intialize
func init() {

	config := base.Configuration{
		EnableConsole:     false,                                 //print output on the console, Good for debugging in local
		ConsoleLevel:      base.Debug,                            //Debug level log
		ConsoleJSONFormat: false,                                 //Console log in JSON format, false will print in raw format on console
		EnableFile:        true,                                  // Logging in File
		FileLevel:         base.Info,                             // File log level
		FileJSONFormat:    true,                                  // File JSON Format, False will print in file in raw Format
		FileLocation:      "/usr/local/production/logs/ctrl.log", //File location where log needs to be appended
	}

	err := base.NewLogger(config)
	if err != nil {
		base.Zlog.Fatalf("Could not instantiate controller log %s", err.Error())
	}
	base.Zlog.Infof("Starting controller logger...")
}

func main() {
	base.Zlog.Infof("Starting controller...")

	err := initCtrlconfig()
	if err != nil {
		base.Zlog.Fatalf("Controller config initialization error: %s", err.Error())
	}
	var ctrlTCPPort = viper.GetString("CTRL_TCPPORT")
	mux := http.NewServeMux()

	// Highest priority must be set to the signed request
	mux.HandleFunc("/", home)

	if err := http.ListenAndServe(ctrlTCPPort, mux); err != nil {
		base.Zlog.Fatalf("Controller service error: %s", err.Error())
	}
}
