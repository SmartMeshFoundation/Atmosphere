package blockchain

import (
	"testing"

	"fmt"

	"time"

	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/codefortest"
	"github.com/SmartMeshFoundation/Atmosphere/network/rpc"
	"github.com/SmartMeshFoundation/Atmosphere/params"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
)

type fakeRPCModule struct {
	TokenNetworkAddress   common.Address
	SecretRegistryAddress common.Address
}

func (r *fakeRPCModule) GetTokenNetworkAddress() common.Address {
	return r.TokenNetworkAddress
}

func (r *fakeRPCModule) GetSecretRegistryAddress() common.Address {
	return r.SecretRegistryAddress
}

func TestNewBlockChainEvents(t *testing.T) {
	client, err := codefortest.GetEthClient()
	if err != nil {
		panic(err)
	}
	be := NewBlockChainEvents(client, &fakeRPCModule{})
	if be == nil {
		t.Error("NewBlockChainEvents failed")
	}
}

func TestEvents_Start(t *testing.T) {
	client, err := codefortest.GetEthClient()
	if err != nil {
		panic(err)
	}
	be := NewBlockChainEvents(client, &fakeRPCModule{
		TokenNetworkAddress: rpc.TestGetTokenNetworkAddress(),
	})
	if be == nil {
		t.Error("NewBlockChainEvents failed")
	}
	params.ChainID = big.NewInt(8888)
	be.Start(-1)
	begin := time.Now()
	for {
		if time.Since(begin) > 10*time.Second {
			be.Stop()
			time.Sleep(5 * time.Second)
			return
		}
		select {
		case sc := <-be.StateChangeChannel:
			fmt.Println(utils.StringInterface(sc, 0))
			//BlockStateChange, ok := sc.(transfer.BlockStateChange)
			//if ok {
			//	fmt.Println(BlockStateChange.BlockNumber)
			//}
		}
	}
}
