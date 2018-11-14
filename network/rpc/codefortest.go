package rpc

import (
	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/network/helper"

	"os"

	"encoding/hex"

	"crypto/ecdsa"

	"github.com/SmartMeshFoundation/Atmosphere/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/encoding"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

//PrivateRopstenRegistryAddress test registry address, todo use env
var PrivateRopstenRegistryAddress = common.HexToAddress(os.Getenv("TOKEN_NETWORK"))

//TestRPCEndpoint test eth rpc url, todo use env
var TestRPCEndpoint = os.Getenv("ETHRPCENDPOINT")

//TestPrivKey for test only
var TestPrivKey *ecdsa.PrivateKey

func init() {
	if encoding.IsTest {
		keybin, err := hex.DecodeString(os.Getenv("KEY1"))
		if err != nil {
			//启动错误不要用 log, 那时候 log 还没准备好
			// do not use log to print start error, it's not ready
			panic(fmt.Sprintf("err %s", err))
		}
		TestPrivKey, err = crypto.ToECDSA(keybin)
		if err != nil {
			panic(fmt.Sprintf("err %s", err))
		}
	}

}

//MakeTestBlockChainService creat test BlockChainService
func MakeTestBlockChainService() *BlockChainService {
	conn, err := helper.NewSafeClient(TestRPCEndpoint)
	//conn, err := ethclient.Dial("ws://" + node.DefaultWSEndpoint())
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the Ethereum client: %s\n", err))
	}
	bcs, err := NewBlockChainService(TestPrivKey, PrivateRopstenRegistryAddress, conn)
	if err != nil {
		panic(err)
	}
	return bcs
}

//GetTestChannelUniqueID for test only,get from env
func GetTestChannelUniqueID() *contracts.ChannelUniqueID {

	cu := &contracts.ChannelUniqueID{
		OpenBlockNumber: 3,
	}
	b, err := hex.DecodeString(os.Getenv("CHANNEL"))
	if err != nil || len(b) != len(cu.ChannelIdentifier) {
		panic("CHANNEL env error")
	}
	copy(cu.ChannelIdentifier[:], b)
	return cu
}

//TestGetTokenNetworkAddress for test only
func TestGetTokenNetworkAddress() common.Address {
	addr := common.HexToAddress(os.Getenv("TOKEN_NETWORK"))
	if addr == utils.EmptyAddress {
		panic("TOKENNETWORK env error")
	}
	log.Trace(fmt.Sprintf("test TOKEN_NETWORK=%s ", addr.String()))
	return addr
}

//TestGetParticipant1 for test only
func TestGetParticipant1() (privKey *ecdsa.PrivateKey, addr common.Address) {
	keybin, err := hex.DecodeString(os.Getenv("KEY1"))
	if err != nil {
		panic("KEY1 ERRor")
	}
	return testGetParticipant(keybin)
}

//TestGetParticipant2 for test only
func TestGetParticipant2() (privKey *ecdsa.PrivateKey, addr common.Address) {
	keybin, err := hex.DecodeString(os.Getenv("KEY2"))
	if err != nil {
		panic("KEY1 ERRor")
	}
	return testGetParticipant(keybin)
}
func testGetParticipant(keybin []byte) (privKey *ecdsa.PrivateKey, addr common.Address) {
	privKey, err := crypto.ToECDSA(keybin)
	if err != nil {
		panic(fmt.Sprintf("toecda err %s,keybin=%s", err, hex.EncodeToString(keybin)))
	}
	addr = crypto.PubkeyToAddress(privKey.PublicKey)
	return
}
