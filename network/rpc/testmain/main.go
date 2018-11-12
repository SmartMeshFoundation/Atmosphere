package main

import (
	"os"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/network/rpc"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, utils.MyStreamHandler(os.Stderr)))
}
func testAddToken() {
	bcs := rpc.MakeTestBlockChainService()
	reg := bcs.Registry(bcs.RegistryProxy.Address, true)
	tokenAddress := common.HexToAddress("0xa9b61a3cc7cc1810e133174caa7ead7ef909d701")
	_, err := reg.AddToken(tokenAddress)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func main() {
	testAddToken()
}
