package rpc

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/network/rpc/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//TokenProxy proxy of ERC20 token
//todo test if support ApproveAndCall ,ERC223 etc
type TokenProxy struct {
	Address common.Address
	bcs     *BlockChainService
	Token   *contracts.Token
}

// TotalSupply total amount of tokens
func (t *TokenProxy) TotalSupply() (*big.Int, error) {
	return t.Token.TotalSupply(t.bcs.getQueryOpts())
}

// BalanceOf The balance
// @param _owner The address from which the balance will be retrieved
func (t *TokenProxy) BalanceOf(addr common.Address) (*big.Int, error) {
	amount, err := t.Token.BalanceOf(t.bcs.getQueryOpts(), addr)
	if err != nil {
		return nil, err
	}
	return amount, err
}

// Allowance Amount of remaining tokens allowed to spent
// @param _owner The address of the account owning tokens
// @param _spender The address of the account able to transfer the tokens
func (t *TokenProxy) Allowance(owner, spender common.Address) (int64, error) {
	amount, err := t.Token.Allowance(t.bcs.getQueryOpts(), owner, spender)
	if err != nil {
		return 0, err
	}
	return amount.Int64(), err //todo if amount larger than max int64?
}

// Approve Whether the approval was successful or not
// @notice `msg.sender` approves `_spender` to spend `_value` tokens
// @param _spender The address of the account able to transfer the tokens
// @param _value The amount of wei to be approved for transfer
func (t *TokenProxy) Approve(spender common.Address, value *big.Int) (err error) {
	tx, err := t.Token.Approve(t.bcs.Auth, spender, value)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Approve %s, txhash=%s", utils.APex(spender), tx.Hash().String()))
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("Approve failed %s,receipt=%s", utils.APex(t.Address), receipt))
		return errors.New("Approve tx execution failed")
	}
	log.Info(fmt.Sprintf("Approve success %s,spender=%s,value=%d", utils.APex(t.Address), utils.APex(spender), value))
	return nil
}

// Transfer return Whether the transfer was successful or not
// @notice send `_value` token to `_to` from `msg.sender`
// @param _to The address of the recipient
// @param _value The amount of token to be transferred
func (t *TokenProxy) Transfer(spender common.Address, value *big.Int) (err error) {
	//由于 abigen Transfer 同名函数生成 bug, 只能先暂时绕开
	err = t.Approve(t.bcs.Auth.From, value)
	if err != nil {
		return
	}
	tx, err := t.Token.TransferFrom(t.bcs.Auth, t.bcs.Auth.From, spender, value)
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("Transfer failed %s,receipt=%s", utils.APex(t.Address), receipt))
		return errors.New("Transfer tx execution failed")
	}
	log.Info(fmt.Sprintf("Transfer success %s,spender=%s,value=%d", utils.APex(t.Address), utils.APex(spender), value))
	return nil
}

//TransferAsync transfer async
func (t *TokenProxy) TransferAsync(spender common.Address, value *big.Int) (result *utils.AsyncResult) {
	result = utils.NewAsyncResult()
	go func() {
		err := t.Transfer(spender, value)
		result.Result <- err
	}()
	return
}

//TransferWithFallback ERC223 TokenFallback
func (t *TokenProxy) TransferWithFallback(to common.Address, value *big.Int, extraData []byte) (err error) {
	tx, err := t.Token.Transfer(t.bcs.Auth, to, value, extraData)
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("TransferWithFallback failed %s,receipt=%s", utils.APex(t.Address), receipt))
		return errors.New("TransferWithFallback tx execution failed")
	}
	log.Info(fmt.Sprintf("TransferWithFallback success %s,spender=%s,value=%d", utils.APex(t.Address), utils.APex(to), value))
	return nil
}

//ApproveAndCall ERC20 extend
func (t *TokenProxy) ApproveAndCall(spender common.Address, value *big.Int, extraData []byte) (err error) {
	tx, err := t.Token.ApproveAndCall(t.bcs.Auth, spender, value, extraData)
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(GetCallContext(), t.bcs.Client, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Info(fmt.Sprintf("ApproveAndCall failed %s,receipt=%s", utils.APex(t.Address), receipt))
		return errors.New("ApproveAndCall tx execution failed")
	}
	log.Info(fmt.Sprintf("ApproveAndCall success %s,spender=%s,value=%d", utils.APex(t.Address), utils.APex(spender), value))
	return nil
}
