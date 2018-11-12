package rpc

import (
	"context"

	"math/big"

	"time"

	"fmt"

	"crypto/ecdsa"

	"sync"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/network/helper"
	"github.com/SmartMeshFoundation/Atmosphere/network/netshare"
	"github.com/SmartMeshFoundation/Atmosphere/network/rpc/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/params"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

//GetCallContext context for tx
func GetCallContext() context.Context {
	ctx, cf := context.WithDeadline(context.Background(), time.Now().Add(params.DefaultTxTimeout))
	if cf != nil {
	}
	return ctx
}

//GetQueryConext context for query on chain
func GetQueryConext() context.Context {
	ctx, cf := context.WithDeadline(context.Background(), time.Now().Add(params.DefaultPollTimeout))
	if cf != nil {
	}
	return ctx
}

/*
BlockChainService provides quering on blockchain.
*/
type BlockChainService struct {
	//PrivKey of this node, todo remove this
	PrivKey *ecdsa.PrivateKey
	//NodeAddress is address of this node
	NodeAddress common.Address
	//RegistryAddress registy contract address
	RegistryProxy       *RegistryProxy
	SecretRegistryProxy *SecretRegistryProxy
	//Client if eth rpc client
	Client          *helper.SafeEthClient
	addressTokens   map[common.Address]*TokenProxy
	addressChannels map[common.Address]*TokenNetworkProxy
	//Auth needs by call on blockchain todo remove this
	Auth *bind.TransactOpts
}

//NewBlockChainService create BlockChainService
func NewBlockChainService(privateKey *ecdsa.PrivateKey, registryAddress common.Address, client *helper.SafeEthClient) (bcs *BlockChainService, err error) {
	bcs = &BlockChainService{
		PrivKey:         privateKey,
		NodeAddress:     crypto.PubkeyToAddress(privateKey.PublicKey),
		Client:          client,
		addressTokens:   make(map[common.Address]*TokenProxy),
		addressChannels: make(map[common.Address]*TokenNetworkProxy),
		Auth:            bind.NewKeyedTransactor(privateKey),
	}
	// remove gas limit config and let it calculate automatically
	//bcs.Auth.GasLimit = uint64(params.GasLimit)
	bcs.Auth.GasPrice = big.NewInt(params.GasPrice)

	bcs.Registry(registryAddress, client.Status == netshare.Connected)
	return bcs, nil
}
func (bcs *BlockChainService) getQueryOpts() *bind.CallOpts {
	return &bind.CallOpts{
		Pending: false,
		From:    bcs.NodeAddress,
		Context: GetQueryConext(),
	}
}
func (bcs *BlockChainService) blockNumber() (num *big.Int, err error) {
	h, err := bcs.Client.HeaderByNumber(context.Background(), nil)
	if err == nil {
		num = h.Number
	}
	return
}
func (bcs *BlockChainService) nonce(account common.Address) (uint64, error) {
	return bcs.Client.PendingNonceAt(context.Background(), account)
}
func (bcs *BlockChainService) balance(account common.Address) (*big.Int, error) {
	return bcs.Client.PendingBalanceAt(context.Background(), account)
}
func (bcs *BlockChainService) contractExist(contractAddr common.Address) bool {
	code, err := bcs.Client.CodeAt(context.Background(), contractAddr, nil)
	//spew.Dump("code:", code)
	return err == nil && len(code) > 0
}
func (bcs *BlockChainService) getBlockHeader(blockNumber *big.Int) (*types.Header, error) {
	return bcs.Client.HeaderByNumber(context.Background(), blockNumber)
}

func (bcs *BlockChainService) nextBlock() (currentBlock *big.Int, err error) {
	currentBlock, err = bcs.blockNumber()
	if err != nil {
		return
	}
	targetBlockNumber := new(big.Int).Add(currentBlock, big.NewInt(1))
	for currentBlock.Cmp(targetBlockNumber) == -1 {
		time.Sleep(500 * time.Millisecond)
		currentBlock, err = bcs.blockNumber()
		if err != nil {
			return
		}
	}
	return
}

// Token return a proxy to interact with a token.
func (bcs *BlockChainService) Token(tokenAddress common.Address) (t *TokenProxy, err error) {
	_, ok := bcs.addressTokens[tokenAddress]
	if !ok {
		token, err := contracts.NewToken(tokenAddress, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewToken %s err %s", tokenAddress.String(), err))
			return nil, err
		}
		bcs.addressTokens[tokenAddress] = &TokenProxy{
			Address: tokenAddress, bcs: bcs, Token: token}
	}
	return bcs.addressTokens[tokenAddress], nil
}

//TokenNetwork return a proxy to interact with a NettingChannelContract.
func (bcs *BlockChainService) TokenNetwork(address common.Address) (t *TokenNetworkProxy, err error) {
	_, ok := bcs.addressChannels[address]
	if !ok {
		var tokenNetwork *contracts.TokenNetwork
		tokenNetwork, err = contracts.NewTokenNetwork(address, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewNettingChannelContract %s err %s", address.String(), err))
			return
		}
		if !bcs.contractExist(address) {
			return nil, fmt.Errorf("no code at %s", address.String())
		}
		bcs.addressChannels[address] = &TokenNetworkProxy{Address: address, bcs: bcs, ch: tokenNetwork}
	}
	return bcs.addressChannels[address], nil
}

//TokenNetworkWithoutCheck return a proxy to interact with a NettingChannelContract,don't check this address is valid or not
func (bcs *BlockChainService) TokenNetworkWithoutCheck(address common.Address) (t *TokenNetworkProxy, err error) {
	_, ok := bcs.addressChannels[address]
	if !ok {
		var ch *contracts.TokenNetwork
		ch, err = contracts.NewTokenNetwork(address, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewNettingChannelContract %s err %s", address.String(), err))
			return
		}
		bcs.addressChannels[address] = &TokenNetworkProxy{Address: address, bcs: bcs, ch: ch}
	}
	return bcs.addressChannels[address], nil
}

// Registry Return a proxy to interact with Registry.
func (bcs *BlockChainService) Registry(address common.Address, hasConnectChain bool) (t *RegistryProxy) {
	if bcs.RegistryProxy != nil && bcs.RegistryProxy.registry != nil {
		return bcs.RegistryProxy
	}
	r := &RegistryProxy{
		Address: address,
		bcs:     bcs,
	}
	if hasConnectChain {
		reg, err := contracts.NewTokenNetworkRegistry(address, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewRegistry %s err %s ", address.String(), err))
			return
		}
		r.registry = reg
		secAddr, err := r.registry.SecretRegistryAddress(nil)
		if err != nil {
			log.Error(fmt.Sprintf("get Secret_registry_address %s", err))
			return
		}
		s, err := contracts.NewSecretRegistry(secAddr, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewSecretRegistry err %s", err))
			return
		}
		bcs.SecretRegistryProxy = &SecretRegistryProxy{
			Address:          secAddr,
			bcs:              bcs,
			registry:         s,
			RegisteredSecret: make(map[common.Hash]*sync.Mutex),
		}
	}
	bcs.RegistryProxy = r
	return bcs.RegistryProxy
}

// GetRegistryAddress :
func (bcs *BlockChainService) GetRegistryAddress() common.Address {
	if bcs.RegistryProxy != nil {
		return bcs.RegistryProxy.Address
	}
	return utils.EmptyAddress
}

// GetSecretRegistryAddress :
func (bcs *BlockChainService) GetSecretRegistryAddress() common.Address {
	if bcs.SecretRegistryProxy != nil {
		return bcs.SecretRegistryProxy.Address
	}
	return utils.EmptyAddress
}
