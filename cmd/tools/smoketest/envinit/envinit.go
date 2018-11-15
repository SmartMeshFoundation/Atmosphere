package main

import (
	"log"

	"context"

	"fmt"

	"os"

	"crypto/ecdsa"

	"math/big"

	"sync"

	"time"

	"github.com/SmartMeshFoundation/Atmosphere/accounts"
	"github.com/SmartMeshFoundation/Atmosphere/cmd/tools/newtestenv/createchannel"
	"github.com/SmartMeshFoundation/Atmosphere/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/contracts/test/tokens/tokenerc223approve"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/huamou/config"
	"gopkg.in/urfave/cli.v1"
)

var globalPassword = "123"
var passwords = []string{"123", "111111", "123456"}
var env, _ = config.ReadDefault("../env.INI")

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "keystore-path",
			Usage: "If you have a non-standard path for the ethereum keystore directory provide it using this argument. ",
			Value: "../../../../testdata/keystore",
		},
		cli.StringFlag{
			Name: "eth-rpc-endpoint",
			Usage: `"host:port" address of ethereum JSON-RPC server.\n'
	           'Also accepts a protocol prefix (ws:// or ipc channel) with optional port',`,
			Value: fmt.Sprintf("http://127.0.0.1:9001"), //, node.DefaultWSEndpoint()),
		},
		cli.BoolFlag{
			Name:  "not-create-channel",
			Usage: "not-create channels between node for test.",
		},
	}
	app.Action = Main
	app.Name = "envinit"
	app.Version = "0.1"
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Main : main
func Main(ctx *cli.Context) error {
	paramsSection := "PHOTON_PARAMS"
	fmt.Printf("eth-rpc-endpoint:%s\n", ctx.String("eth-rpc-endpoint"))
	fmt.Printf("not-create-channel=%v\n", ctx.Bool("not-create-channel"))
	// Create an IPC based RPC connection to a remote node and an authorized transactor
	conn, err := ethclient.Dial(ctx.String("eth-rpc-endpoint"))
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to connect to the Ethereum client: %v", err))
	}

	_, key := promptAccount(ctx.String("keystore-path"))
	fmt.Println("start to deploy ...")
	tokenNetworkAddress := DeployContract(key, conn)
	env.RemoveOption(paramsSection, "token_network_address")
	env.AddOption(paramsSection, "token_network_address", tokenNetworkAddress.String())
	tokenNetwork, err := contracts.NewTokenNetwork(tokenNetworkAddress, conn)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	lock := &sync.Mutex{}
	for i := 0; i < 2; i++ {
		go func() {
			createTokenAndChannels(key, conn, tokenNetworkAddress, tokenNetwork, ctx.String("keystore-path"), !ctx.Bool("not-create-channel"), lock)
			wg.Done()
		}()
	}
	wg.Wait()
	err = env.WriteFile("../env.INI", 0644, "atmosphere smoke test envInit")
	return err
}
func promptAccount(keystorePath string) (addr common.Address, key *ecdsa.PrivateKey) {
	am := accounts.NewAccountManager(keystorePath)
	if len(am.Accounts) == 0 {
		log.Fatal(fmt.Sprintf("No Ethereum accounts found in the directory %s", keystorePath))
		os.Exit(1)
	}
	addr = am.Accounts[0].Address
	log.Printf("deploy account = %s", addr.String())
	log.Printf("accounts=%q", am.Accounts)
	for i := 0; i < 3; i++ {
		//fmt.Printf("\npassword is %s\n", password)
		keybin, err := am.GetPrivateKey(addr, passwords[0])
		if err != nil && i == 3 {
			log.Fatal(fmt.Sprintf("Exhausted passphrase unlock attempts for %s. Aborting ...", addr))
			os.Exit(1)
		}
		if err != nil {
			log.Println(fmt.Sprintf("password incorrect\n Please try again or kill the process to quit.\nUsually Ctrl-c."))
			continue
		}
		key, err = crypto.ToECDSA(keybin)
		if err != nil {
			log.Println(fmt.Sprintf("private key to bytes err %s", err))
		}
		break
	}
	return
}

// DeployContract :
func DeployContract(key *ecdsa.PrivateKey, conn *ethclient.Client) common.Address {
	chainID, err := conn.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("failed get chain id :%s", chainID)
	}
	fmt.Printf("current chain Id=%s\n", chainID)
	auth := bind.NewKeyedTransactor(key)
	//Deploy TokenNetorkRegistry
	//auth.GasLimit = 4000000 //最大gas
	//auth.GasPrice = big.NewInt(2000)
	tokenNetworkAddress, tx, _, err := contracts.DeployTokenNetwork(auth, conn, chainID)
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}
	fmt.Printf("tokenNetworkAddress=%s, txhash=%s\n", tokenNetworkAddress.String(), tx.Hash().String())
	ctx := context.Background()
	_, err = bind.WaitDeployed(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to deploy contact when mining :%v", err)
	}
	fmt.Printf("Deploy tokenNetworkAddress complete...\n")

	fmt.Printf("tokenNetworkAddress=%s\n", tokenNetworkAddress.String())
	return tokenNetworkAddress
}

func createTokenAndChannels(key *ecdsa.PrivateKey, conn *ethclient.Client, tokenNetworkAddress common.Address, tokenNetwork *contracts.TokenNetwork, keystorepath string, createchannel bool, lock *sync.Mutex) {
	lock.Lock()
	tokenAddress := NewToken(key, conn)
	token, err := contracts.NewToken(tokenAddress, conn)
	if err != nil {
		log.Fatalf("err for newtoken err %s", err)
		return
	}
	am := accounts.NewAccountManager(keystorepath)
	var accounts []common.Address
	var keys []*ecdsa.PrivateKey
	for _, account := range am.Accounts {
		accounts = append(accounts, account.Address)
		keybin, err := am.GetPrivateKey(account.Address, globalPassword)
		if err != nil {
			log.Fatalf("password error for %s", account.Address.String())
		}
		keytemp, err := crypto.ToECDSA(keybin)
		if err != nil {
			log.Fatalf("toecdsa err %s", err)
		}
		keys = append(keys, keytemp)
	}
	lock.Unlock()
	fmt.Printf("key=%s", key)
	//createerc20token合约时间较长,导致多个token同时部署的时候Tx nonce会冲突
	time.Sleep(time.Second)
	TransferMoneyForAccounts(key, conn, accounts[1:], token)
	if createchannel {
		CreateChannels(conn, accounts, keys, tokenNetworkAddress, tokenNetwork, tokenAddress, token)
	}
}

// NewToken ：
func NewToken(key *ecdsa.PrivateKey, conn *ethclient.Client) (tokenAddr common.Address) {
	auth := bind.NewKeyedTransactor(key)
	tokenAddr, tx, _, err := tokenerc223approve.DeployHumanERC223Token(auth, conn, big.NewInt(50000000000), "test symoble", 0)
	if err != nil {
		log.Fatalf("Failed to DeployHumanStandardToken: %v", err)
	}
	fmt.Printf("token deploy tx=%s\n", tx.Hash().String())
	ctx := context.Background()
	_, err = bind.WaitDeployed(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to deploy contact when mining :%v", err)
	}
	fmt.Printf("DeployHumanStandardToken complete... token=%s\n", tokenAddr.String())
	return
}

// TransferMoneyForAccounts :
func TransferMoneyForAccounts(key *ecdsa.PrivateKey, conn *ethclient.Client, accounts []common.Address, token *contracts.Token) {
	wg := sync.WaitGroup{}
	wg.Add(len(accounts))
	//auth := bind.NewKeyedTransactor(key)
	//nonce, err := conn.PendingNonceAt(context.Background(), auth.From)
	//if err != nil {
	//	log.Fatal(err)
	//}
	for index, account := range accounts {
		go func(account common.Address, i int) {
			auth2 := bind.NewKeyedTransactor(key)
			//auth2.Nonce = big.NewInt(int64(nonce) + int64(i))
			fmt.Printf("transfer to %s,nonce=%s\n", account.String(), auth2.Nonce)
			var tx *types.Transaction
			tx, err := token.Transfer(auth2, account, big.NewInt(5000000), nil)
			if err != nil {
				log.Fatalf("Failed to Transfer: %v", err)
			}
			ctx := context.Background()
			_, err = bind.WaitMined(ctx, conn, tx)
			if err != nil {
				log.Fatalf("failed to Transfer when mining :%v", err)
			}
			fmt.Printf("Transfer complete...\n")
			wg.Done()
		}(account, index)
		time.Sleep(time.Millisecond * 100)
	}
	wg.Wait()
	for _, account := range accounts {
		b, err := token.BalanceOf(nil, account)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("account %s has token %s\n", utils.APex(account), b)
	}
}

// CreateChannels : path A-B-C-F-B-D-G-E
func CreateChannels(conn *ethclient.Client, accounts []common.Address, keys []*ecdsa.PrivateKey, tokenNetworkAddress common.Address, tokenNetwork *contracts.TokenNetwork, tokenAddress common.Address, token *contracts.Token) {
	if len(accounts) < 6 {
		panic("need 6 accounts")
	}
	AccountA := accounts[0]
	AccountB := accounts[1]
	AccountC := accounts[2]
	AccountD := accounts[3]
	AccountE := accounts[4]
	AccountF := accounts[5]
	AccountG := accounts[6]
	fmt.Printf("accountA=%s\naccountB=%s\naccountC=%s\naccountD=%s\naccountE=%s\naccountF=%s\naccountG=%s\n",
		AccountA.String(), AccountB.String(), AccountC.String(), AccountD.String(),
		AccountE.String(), AccountF.String(), AccountG.String())
	keyA := keys[0]
	keyB := keys[1]
	keyC := keys[2]
	keyD := keys[3]
	keyE := keys[4]
	keyF := keys[5]
	keyG := keys[6]
	wg := sync.WaitGroup{}
	wg.Add(8)
	//fmt.Printf("keya=%s,keyb=%s,keyc=%s,keyd=%s,keye=%s,keyf=%s,keyg=%s", keyA, keyB, keyC, keyD, keyE, keyF, keyG)
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountA, AccountB, keyA, keyB, big.NewInt(100), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountB, AccountD, keyB, keyD, big.NewInt(90), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountB, AccountC, keyB, keyC, big.NewInt(50), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountB, AccountF, keyB, keyF, big.NewInt(70), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountC, AccountF, keyC, keyF, big.NewInt(60), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountC, AccountE, keyC, keyE, big.NewInt(10), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountD, AccountG, keyD, keyG, big.NewInt(90), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	go func() {
		createchannel.CreatAChannelAndDeposit(AccountG, AccountE, keyG, keyE, big.NewInt(80), tokenNetworkAddress, tokenNetwork, tokenAddress, token, conn)
		wg.Done()
	}()
	wg.Wait()
}
