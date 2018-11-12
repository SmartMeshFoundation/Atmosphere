package cases

import (
	"context"
	"crypto/ecdsa"
	"log"
	"net/http"
	"time"

	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/accounts"
	"github.com/SmartMeshFoundation/Atmosphere/cmd/tools/smoketest/models"
	"github.com/SmartMeshFoundation/Atmosphere/network/rpc/contracts/test/tokens/tokenerc223approve"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/huamou/config"
)

// RegisteringTokenTest : test case for register token
func RegisteringTokenTest(env *models.PhotonEnvReader, allowFail bool) {
	// 1. register a not-exist token
	case1 := &APITestCase{
		CaseName:  "Register a not-exist token",
		AllowFail: allowFail,
		Req: &models.Req{
			APIName: "RegisteringOneToken",
			FullURL: env.RandomNode().Host + "/api/1/tokens/0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF",
			Method:  http.MethodPut,
			Payload: "",
			Timeout: time.Second * 120,
		},
		TargetStatusCode: 409,
	}
	case1.Run()
	// 2. register a new token
	newTokenAddress := deployNewToken()
	case2 := &APITestCase{
		CaseName:  "Register a new token",
		AllowFail: allowFail,
		Req: &models.Req{
			APIName: "RegisteringOneToken",
			FullURL: env.RandomNode().Host + "/api/1/tokens/" + newTokenAddress,
			Method:  http.MethodPut,
			Payload: "",
			Timeout: time.Second * 180,
		},
		TargetStatusCode: 200,
	}
	case2.Run()
}

func deployNewToken() (newTokenAddress string) {
	c, err := config.ReadDefault("./env.INI")
	if err != nil {
		log.Println("config.ReadDefault error:", err)
		return
	}
	EthRPCEndpoint := c.RdString("PHOTON_PARAMS", "eth_rpc_endpoint", "ws://127.0.0.1:8546")
	KeyStorePath := c.RdString("PHOTON_PARAMS", "keystore_path", "/smtwork/privnet3/data/keystore")
	conn, err := ethclient.Dial(EthRPCEndpoint)
	if err != nil {
		log.Fatalf("connect to eth rpc error %s", err)
		return
	}
	return deployOneToken(KeyStorePath, conn).String()
}

func deployOneToken(keystorePath string, conn *ethclient.Client) (tokenAddr common.Address) {
	key := getDeployKey(keystorePath)
	auth := bind.NewKeyedTransactor(key)
	tokenAddr, tx, _, err := tokenerc223approve.DeployHumanERC223Token(auth, conn, big.NewInt(50000000000), "test symoble", 0)
	if err != nil {
		log.Fatalf("Failed to DeployHumanStandardToken: %v", err)
	}
	ctx := context.Background()
	_, err = bind.WaitDeployed(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to deploy contact when mining :%v", err)
	}
	return
}

func getDeployKey(keystorePath string) (key *ecdsa.PrivateKey) {
	am := accounts.NewAccountManager(keystorePath)
	if len(am.Accounts) <= 0 {
		log.Fatalf("no accounts @%s", keystorePath)
	}
	keybin, err := am.GetPrivateKey(am.Accounts[0].Address, GlobalPassword)
	if err != nil {
		log.Fatalf("get first private key error %s", err)
		return
	}
	key, err = crypto.ToECDSA(keybin)
	if err != nil {
		log.Fatalf("private key to bytes err  %s", err)
		return
	}
	return
}
