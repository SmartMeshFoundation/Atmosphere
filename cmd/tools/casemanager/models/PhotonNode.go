package models

import (
	"log"
	"time"

	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/SmartMeshFoundation/Atmosphere/params"
)

// PhotonNode a atmosphere node
type PhotonNode struct {
	Host          string
	Address       string
	Name          string
	APIAddress    string
	ListenAddress string
	ConditionQuit *params.ConditionQuit
	DebugCrash    bool
	Running       bool
}

// Start start a atmosphere node
func (node *PhotonNode) Start(env *TestEnv) {
	logfile := fmt.Sprintf("./log/%s.log", env.CaseName+"-"+node.Name)
	go ExecShell(env.Main, node.getParamStr(env), logfile, true)

	count := 0
	t := time.Now()
	for !node.IsRunning() {
		Logger.Printf("waiting for %s to start, sleep 3s...\n", node.Name)
		time.Sleep(time.Second * 3)
		count++
		if count > 40 {
			if node.ConditionQuit != nil {
				Logger.Printf("NODE %s %s start with %s TIMEOUT\n", node.Address, node.Host, node.ConditionQuit.QuitEvent)
			} else {
				Logger.Printf("NODE %s %s start TIMEOUT\n", node.Address, node.Host)
			}
			panic("Start atmosphere node TIMEOUT")
		}
	}
	used := time.Since(t)
	if node.DebugCrash {
		Logger.Printf("NODE %s %s start with %s in %fs", node.Address, node.Host, node.ConditionQuit.QuitEvent, used.Seconds())
	} else {
		Logger.Printf("NODE %s %s start in %fs", node.Address, node.Host, used.Seconds())
	}
	time.Sleep(10 * time.Second)
	node.Running = true
	for _, n := range env.Nodes {
		if n.Running {
			n.UpdateMeshNetworkNodes(env.Nodes...)
		}
	}
}

// StartWithParams start a atmosphere node with --fee
func (node *PhotonNode) StartWithParams(env *TestEnv, otherParams ...string) {
	logfile := fmt.Sprintf("./log/%s.log", env.CaseName+"-"+node.Name)
	params := node.getParamStrWithoutNoNetwork(env)
	params = append(params, otherParams...)
	go ExecShell(env.Main, params, logfile, true)

	count := 0
	t := time.Now()
	for !node.IsRunning() {
		Logger.Printf("waiting for %s to StartWithParams, sleep 3s...\n", node.Name)
		time.Sleep(time.Second * 3)
		count++
		if count > 40 {
			if node.ConditionQuit != nil {
				Logger.Printf("NODE %s %s StartWithParams with %s TIMEOUT\n", node.Address, node.Host, node.ConditionQuit.QuitEvent)
			} else {
				Logger.Printf("NODE %s %s StartWithParams TIMEOUT\n", node.Address, node.Host)
			}
			panic("Start atmosphere node TIMEOUT")
		}
	}
	used := time.Since(t)
	if node.DebugCrash {
		Logger.Printf("NODE %s %s StartWithParams with %s in %fs", node.Address, node.Host, node.ConditionQuit.QuitEvent, used.Seconds())
	} else {
		Logger.Printf("NODE %s %s StartWithParams in %fs", node.Address, node.Host, used.Seconds())
	}
	time.Sleep(10 * time.Second)
	node.Running = true
}

// StartWithFee start a atmosphere node with --fee
func (node *PhotonNode) StartWithFee(env *TestEnv) {
	logfile := fmt.Sprintf("./log/%s.log", env.CaseName+"-"+node.Name)
	params := node.getParamStr(env)
	params = append(params, "--fee")
	go ExecShell(env.Main, params, logfile, true)

	count := 0
	t := time.Now()
	for !node.IsRunning() {
		Logger.Printf("waiting for %s to StartWithFee, sleep 3s...\n", node.Name)
		time.Sleep(time.Second * 3)
		count++
		if count > 40 {
			if node.ConditionQuit != nil {
				Logger.Printf("NODE %s %s StartWithFee with %s TIMEOUT\n", node.Address, node.Host, node.ConditionQuit.QuitEvent)
			} else {
				Logger.Printf("NODE %s %s StartWithFee TIMEOUT\n", node.Address, node.Host)
			}
			panic("Start atmosphere node TIMEOUT")
		}
	}
	used := time.Since(t)
	if node.DebugCrash {
		Logger.Printf("NODE %s %s StartWithFee with %s in %fs", node.Address, node.Host, node.ConditionQuit.QuitEvent, used.Seconds())
	} else {
		Logger.Printf("NODE %s %s StartWithFee in %fs", node.Address, node.Host, used.Seconds())
	}
	time.Sleep(10 * time.Second)
	node.Running = true
	for _, n := range env.Nodes {
		if n.Running {
			n.UpdateMeshNetworkNodes(env.Nodes...)
		}
	}
}

// ReStartWithoutConditionquit : Restart start a atmosphere node
func (node *PhotonNode) ReStartWithoutConditionquit(env *TestEnv) {
	node.DebugCrash = false
	node.ConditionQuit = nil
	node.Name = node.Name + "Restart"
	node.Start(env)
}

func (node *PhotonNode) getParamStr(env *TestEnv) []string {
	var param []string
	param = append(param, "--datadir="+env.DataDir)
	param = append(param, "--api-address="+node.APIAddress)
	param = append(param, "--listen-address="+node.ListenAddress)
	param = append(param, "--address="+node.Address)
	param = append(param, "--keystore-path="+env.KeystorePath)
	param = append(param, "--token-network-address="+env.TokenNetworkAddress)
	param = append(param, "--password-file="+env.PasswordFile)
	param = append(param, "--nonetwork")
	if env.XMPPServer != "" {
		param = append(param, "--xmpp-server="+env.XMPPServer)
	}
	param = append(param, "--eth-rpc-endpoint="+env.EthRPCEndpoint)
	param = append(param, fmt.Sprintf("--verbosity=%d", env.Verbosity))
	if env.Debug == true {
		param = append(param, "--debug")
	}
	if node.DebugCrash == true {
		buf, err := json.Marshal(node.ConditionQuit)
		if err != nil {
			panic(err)
		}
		param = append(param, "--debugcrash")
		param = append(param, "--conditionquit="+string(buf))
	}
	return param
}

func (node *PhotonNode) getParamStrWithoutNoNetwork(env *TestEnv) []string {
	var param []string
	param = append(param, "--datadir="+env.DataDir)
	param = append(param, "--api-address="+node.APIAddress)
	param = append(param, "--listen-address="+node.ListenAddress)
	param = append(param, "--address="+node.Address)
	param = append(param, "--keystore-path="+env.KeystorePath)
	param = append(param, "--token_network_address="+env.TokenNetworkAddress)
	param = append(param, "--password-file="+env.PasswordFile)
	if env.XMPPServer != "" {
		param = append(param, "--xmpp-server="+env.XMPPServer)
	}
	param = append(param, "--eth-rpc-endpoint="+env.EthRPCEndpoint)
	param = append(param, fmt.Sprintf("--verbosity=%d", env.Verbosity))
	if env.Debug == true {
		param = append(param, "--debug")
	}
	if node.DebugCrash == true {
		buf, err := json.Marshal(node.ConditionQuit)
		if err != nil {
			panic(err)
		}
		param = append(param, "--debugcrash")
		param = append(param, "--conditionquit="+string(buf))
	}
	return param
}

// StartWithConditionQuit start a atmosphere node whit condition quit
func (node *PhotonNode) StartWithConditionQuit(env *TestEnv, c *params.ConditionQuit) {
	node.ConditionQuit = c
	node.DebugCrash = true
	node.Start(env)
}

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
