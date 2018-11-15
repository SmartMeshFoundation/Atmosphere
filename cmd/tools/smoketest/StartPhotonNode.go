package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fmt"
	"runtime"
	"strconv"

	"github.com/huamou/config"
	"github.com/kataras/iris/utils"
)

// ExecShell : run shell commands
func ExecShell(cmdstr string, param []string, logfile string, canquit bool) bool {
	var err error
	/* #nosec */
	cmd := exec.Command(cmdstr, param...)

	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return false
	}

	reader := bufio.NewReader(stdout)
	readererr := bufio.NewReader(stderr)

	logPath := filepath.Dir(logfile)
	if !utils.Exists(logPath) {
		err = os.Mkdir(logPath, 0700)
	}

	logFile, err := os.Create(logfile)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("Create log file error !", logfile)
	}

	debugLog := log.New(logFile, "", 0)
	//A real-time loop reads a line in the output stream.
	go func() {
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil || io.EOF == readErr {
				break
			}
			//log.Println(line)
			debugLog.Println(line)
		}
	}()

	//go func() {
	for {
		line, readErr := readererr.ReadString('\n')
		if readErr != nil || io.EOF == readErr {
			break
		}
		//log.Println(line)
		debugLog.Println(line)
	}
	//}()

	err = cmd.Wait()

	if !canquit {
		log.Println("cmd ", cmdstr, " exited:", param)
	}

	if err != nil {
		//log.Println(err)
		debugLog.Println(err)
		if !canquit {
			os.Exit(-1)
		}
		return false
	}
	if !canquit {
		os.Exit(-1)
	}
	return true
}

// StartPhotonNode : start atmosphere
func StartPhotonNode(TokenNetworkAddress string) {
	paramsSection := "PHOTON_PARAMS"
	var pstr []string
	//public parameter
	var pstr2 []string
	//kill the old process
	if runtime.GOOS == "windows" {
		pstr2 = append(pstr2, "-F")
		pstr2 = append(pstr2, "-IM")
		pstr2 = append(pstr2, "atmosphere*")
		ExecShell("taskkill", pstr2, "./log/killall.log", true)
	} else {
		pstr2 = append(pstr2, "-s INT")
		pstr2 = append(pstr2, "atmosphere")
		ExecShell("killall", pstr2, "./log/killall.log", true)
	}
	//kill the old process and wait for the release of the port
	time.Sleep(1 * time.Second)

	// get the params
	param := new(PhotonParam)
	c, err := config.ReadDefault("./env.INI")
	if err != nil {
		log.Println("Read error:", err)
		return
	}
	param.datadir = c.RdString(paramsSection, "datadir", "/smtwork/share/.atmosphere")
	param.keystorePath = c.RdString(paramsSection, "keystore_path", "/smtwork/privnet3/data/keystore")
	if TokenNetworkAddress == "" {
		param.tokenNetworkAddress = c.RdString(paramsSection, "token_network_address", "")

	} else {
		param.tokenNetworkAddress = TokenNetworkAddress
	}

	param.passwordFile = c.RdString(paramsSection, "password_file", "")
	param.ethRPCEndpoint = c.RdString(paramsSection, "eth_rpc_endpoint", "ws://127.0.0.1:8546")
	param.debug = c.RdBool(paramsSection, "debug", true)
	//start 6 atmosphere node
	var NODE string
	exepath := c.RdString(paramsSection, "atmosphere_path", "")
	for i := 0; i < 6; i++ {
		NODE = "N" + strconv.Itoa(i)
		param.apiAddress = "0.0.0.0:60" + fmt.Sprintf("%02d", i)
		param.listenAddress = "127.0.0.1:600" + fmt.Sprintf("%02d", i)
		param.address = c.RdString("ACCOUNT", NODE, "")
		pstr = param.getParam()
		logfile := fmt.Sprintf("./log/N%d.log", i)
		go ExecShell(exepath, pstr, logfile, false)
	}
	log.Println("Sleep 60 seconds to wait atmosphere nodes start ...")
	time.Sleep(60 * time.Second)
	log.Println("Atmosphere nodes start done")
}
