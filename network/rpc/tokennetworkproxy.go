package rpc

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/transfer/mtree"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//TokenNetworkProxy proxy of TokenNetwork Contract
type TokenNetworkProxy struct {
	Address common.Address //this contract address

	bcs      *BlockChainService
	contract *contracts.TokenNetwork
}

func to32bytes(src []byte) []byte {
	dst := common.BytesToHash(src)
	return dst[:]
}

func makeDepositData(participantAddress, partnerAddress common.Address, settleTimeout int) []byte {
	var err error
	buf := new(bytes.Buffer)
	_, err = buf.Write(to32bytes(participantAddress[:]))
	_, err = buf.Write(to32bytes(partnerAddress[:]))
	_, err = buf.Write(utils.BigIntTo32Bytes(big.NewInt(int64(settleTimeout)))) //settle_timeout
	if err != nil {
		log.Error(fmt.Sprintf("buf write err %s", err))
	}
	return buf.Bytes()
}
func (t *TokenNetworkProxy) depositByFallback(tokenProxy *TokenProxy, participant, partner common.Address, amount *big.Int, settleTimeout int) (err error) {
	data := makeDepositData(participant, partner, settleTimeout)
	return tokenProxy.TransferWithFallback(t.Address, amount, data)
}
func (t *TokenNetworkProxy) depositByApproveAndCall(tokenProxy *TokenProxy, participant, partner common.Address, amount *big.Int, settleTimeout int) (err error) {
	data := makeDepositData(participant, partner, settleTimeout)
	return tokenProxy.ApproveAndCall(t.Address, amount, data)
}
func (t *TokenNetworkProxy) depositByApprove(tokenProxy *TokenProxy, participant, partner common.Address, amount *big.Int, settleTimeout int) (err error) {
	err = tokenProxy.Approve(t.Address, amount)
	if err != nil {
		return
	}
	tx, err := t.contract.Deposit(t.bcs.Auth, tokenProxy.Address, participant, partner, amount, uint64(settleTimeout))
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> Deposit  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Warn(fmt.Sprintf("ContractCall -> Deposit failed %s", receipt))
		return errors.New("ContractCall -> Deposit tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> Deposit success %s ", utils.APex(t.Address)))
	return nil
}

//Deposit  to  a channel
func (t *TokenNetworkProxy) Deposit(tokenAddress, participant, partner common.Address, amount *big.Int, settleTimeout int) (err error) {
	token, err := t.bcs.NewTokenProxy(tokenAddress)
	if err != nil {
		return
	}
	err = t.depositByFallback(token, participant, partner, amount, settleTimeout)
	if err == nil {
		log.Trace(fmt.Sprintf("ContractCall -> %s-%s depositByFallback success", utils.APex(tokenAddress), utils.APex(partner)))
		return
	}
	err = t.depositByApproveAndCall(token, participant, partner, amount, settleTimeout)
	if err == nil {
		log.Trace(fmt.Sprintf("ContractCall -> %s-%s depositByApproveAndCall success", utils.APex(tokenAddress), utils.APex(partner)))
		return
	}
	return t.depositByApprove(token, participant, partner, amount, settleTimeout)
}

//DepositAsync to  a channel async
func (t *TokenNetworkProxy) DepositAsync(tokenAddress, participant, partner common.Address, amount *big.Int, settleTimeout int) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.Deposit(tokenAddress, participant, partner, amount, settleTimeout)
		result.Result <- err
	}()
	return
}

/*GetChannelInfo Returns the channel specific data.
@param participant1 Address of one of the channel participants.
@param participant2 Address of the other channel participant.
@return ch state and settle_block_number.
if state is 1, settleBlockNumber is settle timeout, if state is 2,settleBlockNumber is the min block number ,settle can be called.
*/
func (t *TokenNetworkProxy) GetChannelInfo(tokenAddress, participant1, participant2 common.Address) (channelID common.Hash, settleBlockNumber, openBlockNumber uint64, state uint8, settleTimeout uint64, err error) {
	return t.contract.GetChannelInfo(t.bcs.getQueryOpts(), tokenAddress, participant1, participant2)
}

//GetChannelParticipantInfo Returns Info of this channel.
//@return The address of the token.
func (t *TokenNetworkProxy) GetChannelParticipantInfo(tokenAddress, participant, partner common.Address) (deposit *big.Int, balanceHash common.Hash, nonce uint64, err error) {
	deposit, h, nonce, err := t.contract.GetChannelParticipantInfo(t.bcs.getQueryOpts(), tokenAddress, participant, partner)
	balanceHash = common.BytesToHash(h[:])
	return
}

//CloseChannel close channel
func (t *TokenNetworkProxy) CloseChannel(tokenAddress, partnerAddr common.Address, transferAmount *big.Int, locksRoot common.Hash, nonce uint64, extraHash common.Hash, signature []byte) (err error) {
	tx, err := t.contract.PrepareSettle(t.bcs.Auth, tokenAddress, partnerAddr, transferAmount, locksRoot, uint64(nonce), extraHash, signature)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> CloseChannel  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("ContractCall -> CloseChannel failed %s", receipt))
		return errors.New("ContractCall -> CloseChannel tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> CloseChannel success %s ,partner=%s", utils.APex(t.Address), utils.APex(partnerAddr)))
	return nil
}

//CloseChannelAsync close channel async
func (t *TokenNetworkProxy) CloseChannelAsync(tokenAddress, partnerAddr common.Address, transferAmount *big.Int, locksRoot common.Hash, nonce uint64, extraHash common.Hash, signature []byte) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.CloseChannel(tokenAddress, partnerAddr, transferAmount, locksRoot, nonce, extraHash, signature)
		result.Result <- err
	}()
	return
}

//UpdateBalanceProof update balance proof of partner
func (t *TokenNetworkProxy) UpdateBalanceProof(tokenAddress, partnerAddr common.Address, transferAmount *big.Int, locksRoot common.Hash, nonce uint64, extraHash common.Hash, signature []byte) (err error) {
	tx, err := t.contract.UpdateBalanceProof(t.bcs.Auth, tokenAddress, partnerAddr, transferAmount, locksRoot, nonce, extraHash, signature)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> UpdateBalanceProof  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("ContractCall -> UpdateBalanceProof failed %s", receipt))
		return errors.New("ContractCall -> UpdateBalanceProof tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> UpdateBalanceProof success %s ,partner=%s", utils.APex(t.Address), utils.APex(partnerAddr)))
	return nil
}

//UpdateBalanceProofAsync update balance proof async
func (t *TokenNetworkProxy) UpdateBalanceProofAsync(tokenAddress, partnerAddr common.Address, transferAmount *big.Int, locksRoot common.Hash, nonce uint64, extraHash common.Hash, signature []byte) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.UpdateBalanceProof(tokenAddress, partnerAddr, transferAmount, locksRoot, nonce, extraHash, signature)
		result.Result <- err
	}()

	return
}

//Unlock a partner's lock
func (t *TokenNetworkProxy) Unlock(tokenAddress, partnerAddr common.Address, transferAmount *big.Int, lock *mtree.Lock, proof []byte) (err error) {
	tx, err := t.contract.Unlock(t.bcs.Auth, tokenAddress, partnerAddr, transferAmount, big.NewInt(lock.Expiration), lock.Amount, lock.LockSecretHash, proof)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> Unlock  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("ContractCall -> Unlock failed %s", receipt))
		return errors.New("ContractCall -> Unlock tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> Unlock success %s ,partner=%s", utils.APex(t.Address), utils.APex(partnerAddr)))
	return nil
}

//UnlockAsync a partner's lock async
func (t *TokenNetworkProxy) UnlockAsync(tokenAddress, partnerAddr common.Address, transferAmount *big.Int, lock *mtree.Lock, proof []byte) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.Unlock(tokenAddress, partnerAddr, transferAmount, lock, proof)
		result.Result <- err
	}()
	return
}

//SettleChannel settle a channel
func (t *TokenNetworkProxy) SettleChannel(tokenAddress, p1Addr, p2Addr common.Address, p1Amount, p2Amount *big.Int, p1Locksroot, p2Locksroot common.Hash) (err error) {
	tx, err := t.contract.Settle(t.bcs.Auth, tokenAddress, p1Addr, p1Amount, p1Locksroot, p2Addr, p2Amount, p2Locksroot)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> SettleChannel  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Warn(fmt.Sprintf("ContractCall -> SettleChannel failed %s", receipt))
		return errors.New("ContractCall -> SettleChannel tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> SettleChannel success %s ", utils.APex(t.Address)))
	return nil
}

//SettleChannelAsync settle a channel async
func (t *TokenNetworkProxy) SettleChannelAsync(tokenAddress, p1Addr, p2Addr common.Address, p1Amount, p2Amount *big.Int, p1Locksroot, p2Locksroot common.Hash) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.SettleChannel(tokenAddress, p1Addr, p2Addr, p1Amount, p2Amount, p1Locksroot, p2Locksroot)
		result.Result <- err
	}()
	return
}

//Withdraw  to  a channel
func (t *TokenNetworkProxy) Withdraw(tokenAddress, p1Addr, p2Addr common.Address, p1Balance,
	p1Withdraw *big.Int, p1Signature, p2Signature []byte) (err error) {
	tx, err := t.contract.WithDraw(t.bcs.Auth, tokenAddress, p1Addr, p2Addr, p1Balance, p1Withdraw,
		p1Signature, p2Signature,
	)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> Withdraw  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Warn(fmt.Sprintf("ContractCall -> Withdraw failed %s", receipt))
		return errors.New("ContractCall -> Withdraw tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> Withdraw success %s ", utils.APex(t.Address)))
	return nil
}

//WithdrawAsync   a channel async
func (t *TokenNetworkProxy) WithdrawAsync(tokenAddress, p1Addr, p2Addr common.Address, p1Balance,
	p1Withdraw *big.Int, p1Signature, p2Signature []byte) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.Withdraw(tokenAddress, p1Addr, p2Addr, p1Balance, p1Withdraw, p1Signature, p2Signature)
		result.Result <- err
	}()
	return
}

//PunishObsoleteUnlock  to  a channel
func (t *TokenNetworkProxy) PunishObsoleteUnlock(tokenAddress, beneficiary, cheater common.Address, lockhash, extraHash common.Hash, cheaterSignature []byte) (err error) {
	tx, err := t.contract.PunishObsoleteUnlock(t.bcs.Auth, tokenAddress, beneficiary, cheater, lockhash, extraHash, cheaterSignature)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> PunishObsoleteUnlock  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Warn(fmt.Sprintf("ContractCall -> PunishObsoleteUnlock failed %s", receipt))
		return errors.New("ContractCall -> PunishObsoleteUnlock tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> PunishObsoleteUnlock success %s ", utils.APex(t.Address)))
	return nil
}

//PunishObsoleteUnlockAsync   a channel async
func (t *TokenNetworkProxy) PunishObsoleteUnlockAsync(tokenAddress, beneficiary, cheater common.Address, lockhash, extraHash common.Hash, cheaterSignature []byte) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.PunishObsoleteUnlock(tokenAddress, beneficiary, cheater, lockhash, extraHash, cheaterSignature)
		result.Result <- err
	}()
	return
}

//CooperativeSettle  settle  a channel
func (t *TokenNetworkProxy) CooperativeSettle(tokenAddress, p1Addr, p2Addr common.Address, p1Balance, p2Balance *big.Int, p1Signature, p2Signatue []byte) (err error) {
	tx, err := t.contract.CooperativeSettle(t.bcs.Auth, tokenAddress, p1Addr, p1Balance, p2Addr, p2Balance, p1Signature, p2Signatue)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("ContractCall -> CooperativeSettle  txhash=%s", tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Warn(fmt.Sprintf("ContractCall -> CooperativeSettle failed %s", receipt))
		return errors.New("ContractCall -> CooperativeSettle tx execution failed")
	}
	log.Info(fmt.Sprintf("ContractCall -> CooperativeSettle success %s ", utils.APex(t.Address)))
	return nil
}

//CooperativeSettleAsync  settle  a channel async
func (t *TokenNetworkProxy) CooperativeSettleAsync(tokenAddress, p1Addr, p2Addr common.Address, p1Balance, p2Balance *big.Int, p1Signature, p2Signatue []byte) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.CooperativeSettle(tokenAddress, p1Addr, p2Addr, p1Balance, p2Balance, p1Signature, p2Signatue)
		result.Result <- err
	}()
	return
}

//GetContractVersion  :
func (t *TokenNetworkProxy) GetContractVersion() (version string, err error) {
	if t.contract == nil {
		err = errors.New("not connected")
		return
	}
	return t.contract.ContractVersion(nil)
}
