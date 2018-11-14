package rpc

import (
	"context"

	"math/big"

	"time"

	"crypto/ecdsa"

	"fmt"
	"sync"

	"github.com/SmartMeshFoundation/Atmosphere/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/network/helper"
	"github.com/SmartMeshFoundation/Atmosphere/network/netshare"
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
	NodeAddress         common.Address
	TokenNetworkProxy   *TokenNetworkProxy
	SecretRegistryProxy *SecretRegistryProxy
	//Client if eth rpc client
	Client          *helper.SafeEthClient
	addressTokens   map[common.Address]*TokenProxy
	addressChannels map[common.Address]*TokenNetworkProxy
	//Auth needs by call on blockchain todo remove this
	Auth *bind.TransactOpts
}

//NewBlockChainService create BlockChainService
func NewBlockChainService(privateKey *ecdsa.PrivateKey, tokenNetworkAddress common.Address, client *helper.SafeEthClient) (bcs *BlockChainService, err error) {
	bcs = &BlockChainService{
		PrivKey:         privateKey,
		NodeAddress:     crypto.PubkeyToAddress(privateKey.PublicKey),
		Client:          client,
		addressTokens:   make(map[common.Address]*TokenProxy),
		addressChannels: make(map[common.Address]*TokenNetworkProxy),
		Auth:            bind.NewKeyedTransactor(privateKey),
	}
	// remove gas limit config and let it calculate automatically
	//bcs.Auth.DefaultGasLimit = uint64(params.DefaultGasLimit)
	bcs.Auth.GasPrice = big.NewInt(params.DefaultGasPrice)

	bcs.NewTokenNetworkProxy(tokenNetworkAddress, client.Status == netshare.Connected)
	return bcs, nil
}

// NewTokenProxy return a proxy to interact with a token.
func (bcs *BlockChainService) NewTokenProxy(tokenAddress common.Address) (proxy *TokenProxy, err error) {
	_, ok := bcs.addressTokens[tokenAddress]
	if !ok {
		token, err := contracts.NewToken(tokenAddress, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewTokenProxy %s err %s", tokenAddress.String(), err))
			return nil, err
		}
		bcs.addressTokens[tokenAddress] = &TokenProxy{
			Address: tokenAddress, bcs: bcs, contract: token}
	}
	return bcs.addressTokens[tokenAddress], nil
}

// NewTokenNetworkProxy Return a proxy to interact with Registry.
func (bcs *BlockChainService) NewTokenNetworkProxy(address common.Address, hasConnectChain bool) (t *TokenNetworkProxy) {
	if bcs.TokenNetworkProxy != nil {
		return bcs.TokenNetworkProxy
	}
	r := &TokenNetworkProxy{
		Address: address,
		bcs:     bcs,
	}
	if hasConnectChain {
		tokenNetwork, err := contracts.NewTokenNetwork(address, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewRegistry %s err %s ", address.String(), err))
			return
		}
		r.contract = tokenNetwork
		secAddr, err := r.contract.SecretRegistry(nil)
		if err != nil {
			log.Error(fmt.Sprintf("get secret_registry_address %s", err))
			return
		}
		s, err := contracts.NewSecretRegistry(secAddr, bcs.Client)
		if err != nil {
			log.Error(fmt.Sprintf("NewSecretRegistry err %s", err))
			return
		}
		bcs.SecretRegistryProxy = &SecretRegistryProxy{
			Address:         secAddr,
			bcs:             bcs,
			contract:        s,
			registryLockMap: make(map[common.Hash]*sync.Mutex),
		}
	}
	bcs.TokenNetworkProxy = r
	return bcs.TokenNetworkProxy
}

// GetTokenNetworkAddress :
func (bcs *BlockChainService) GetTokenNetworkAddress() common.Address {
	if bcs.TokenNetworkProxy != nil {
		return bcs.TokenNetworkProxy.Address
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

// IsConnected  :
func (bcs *BlockChainService) IsConnected() bool {
	return bcs.TokenNetworkProxy != nil && bcs.TokenNetworkProxy.contract != nil && bcs.Client.Status == netshare.Connected
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
