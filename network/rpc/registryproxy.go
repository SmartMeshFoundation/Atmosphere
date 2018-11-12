package rpc

import (
	"errors"
	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/network/rpc/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/rerr"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//RegistryProxy proxy for registry contract
type RegistryProxy struct {
	Address  common.Address //contract address
	bcs      *BlockChainService
	registry *contracts.TokenNetworkRegistry
}

// TokenNetworkByToken Get the ChannelManager address for a specific token
// @param token_address The address of the given token
// @return Address of tokenNetwork
func (r *RegistryProxy) TokenNetworkByToken(tokenAddress common.Address) (tokenNetworkAddress common.Address, err error) {
	if r.registry == nil {
		err = errors.New("registry does't init")
		return
	}
	tokenNetworkAddress, err = r.registry.TokenToTokenNetworks(r.bcs.getQueryOpts(), tokenAddress)
	if tokenNetworkAddress == utils.EmptyAddress {
		err = rerr.ErrNoTokenManager
	}
	return
}

//GetContract return contract itself
func (r *RegistryProxy) GetContract() *contracts.TokenNetworkRegistry {
	return r.registry
}

// GetContractVersion :
func (r *RegistryProxy) GetContractVersion() (contractVersion string, err error) {
	if r.registry == nil {
		err = errors.New("registry does't init")
		return
	}
	return r.registry.ContractVersion(r.bcs.getQueryOpts())
}

//AddToken register a new token,this token must be a valid erc20
func (r *RegistryProxy) AddToken(tokenAddress common.Address) (tokenNetworkAddress common.Address, err error) {
	if r.registry == nil {
		err = errors.New("registry does't init")
		return
	}
	tx, err := r.registry.CreateERC20TokenNetwork(r.bcs.Auth, tokenAddress)
	if err != nil {
		return
	}
	receipt, err := bind.WaitMined(GetCallContext(), r.bcs.Client, tx)
	if err != nil {
		return
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("AddToken failed %s,receipt=%s", utils.APex(r.Address), receipt))
		err = errors.New("AddToken tx execution failed")
		return
	}
	log.Info(fmt.Sprintf("AddToken success %s,token=%s, gasused=%d\n", utils.APex(r.Address), tokenAddress.String(), receipt.GasUsed))
	return r.registry.TokenToTokenNetworks(r.bcs.getQueryOpts(), tokenAddress)
}
