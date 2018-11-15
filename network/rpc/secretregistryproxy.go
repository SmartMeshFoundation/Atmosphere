package rpc

import (
	"errors"
	"fmt"
	"sync"

	"github.com/SmartMeshFoundation/Atmosphere/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//SecretRegistryProxy proxy of secret registry
type SecretRegistryProxy struct {
	Address common.Address

	bcs             *BlockChainService
	contract        *contracts.SecretRegistry
	lock            sync.Mutex
	registryLockMap map[common.Hash]*sync.Mutex
}

//RegisterSecret register secret on chain 有可能被重复调用,但是保证不会并发注册同一个密码
// RegisterSecret : function to register a secret on-chain.
// This function can be repeatedly invoked, and ensure that there is no case that the same secret can be registered concurrently.
func (s *SecretRegistryProxy) RegisterSecret(secret common.Hash) (err error) {
	s.lock.Lock()
	sp := s.registryLockMap[secret]
	if sp == nil {
		sp = &sync.Mutex{}
		s.registryLockMap[secret] = sp
	}
	s.lock.Unlock()
	sp.Lock()
	defer sp.Unlock()
	log.Trace(fmt.Sprintf("ContractCall -> RegisterSecret %s on chain", secret.String()))
	block, err := s.contract.GetSecretRevealBlockHeight(nil, utils.ShaSecret(secret[:]))
	if err == nil && block.Uint64() > 0 {
		//已经注册过了,直接报错
		err = fmt.Errorf("ContractCall -> secret %s,secret hash=%s  already registered", secret.String(), utils.ShaSecret(secret[:]).String())
		return
	}
	tx, err := s.contract.RegisterSecret(s.bcs.Auth, secret)
	if err != nil {
		return err
	}
	log.Trace(fmt.Sprintf("ContractCall -> RegisterSecret on chain tx=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), s.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("ContractCall -> RegisterSecret failed %s,receipt=%s", utils.HPex(secret), receipt))
		return errors.New("ContractCall -> RegisterSecret tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> RegisterSecret success %s,secret=%s", utils.HPex(secret), secret.String()))
	return nil
}

//RegisterSecretAsync 异步注册一个密码
// RegisterSecretAsync : function to register a secret asynchronously.
func (s *SecretRegistryProxy) RegisterSecretAsync(secret common.Hash) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := s.RegisterSecret(secret)
		result.Result <- err
	}()
	return result
}

//IsSecretRegistered 密码是否在合约上注册过,注册地址对不对
// IsSecretRegistered : function to check whether this secret has been registered on chain, and whether the address is correct
func (s *SecretRegistryProxy) IsSecretRegistered(secret common.Hash) (bool, error) {
	blockNumber, err := s.contract.GetSecretRevealBlockHeight(nil, utils.ShaSecret(secret[:]))
	if err != nil {
		return false, err
	}
	if blockNumber.Cmp(utils.BigInt0) <= 0 {
		return false, nil
	}
	return true, nil
}