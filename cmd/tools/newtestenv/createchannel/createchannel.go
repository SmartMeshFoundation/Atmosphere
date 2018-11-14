package createchannel

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"encoding/hex"

	"github.com/SmartMeshFoundation/Atmosphere/contracts"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

//TransferTo ether to address
func TransferTo(conn *ethclient.Client, from *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()
	auth := bind.NewKeyedTransactor(from)
	fromaddr := auth.From
	nonce, err := conn.NonceAt(ctx, fromaddr, nil)
	if err != nil {
		return err
	}
	msg := ethereum.CallMsg{From: fromaddr, To: &to, Value: amount, Data: nil}
	gasLimit, err := conn.EstimateGas(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	gasPrice, err := conn.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %v", err)
	}
	rawTx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	// Create the transaction, sign it and schedule it for execution

	signedTx, err := auth.Signer(types.HomesteadSigner{}, auth.From, rawTx)
	if err != nil {
		return err
	}
	if err = conn.SendTransaction(ctx, signedTx); err != nil {
		return err
	}
	_, err = bind.WaitMined(ctx, conn, signedTx)
	if err != nil {
		return err
	}
	fmt.Printf("transfer from %s to %s amount=%s\n", fromaddr.String(), to.String(), amount)
	return nil
}

//CreatAChannelAndDeposit create a channel
func CreatAChannelAndDeposit(account1, account2 common.Address, key1, key2 *ecdsa.PrivateKey, amount *big.Int, tokenNetworkAddress common.Address, tokenNetwork *contracts.TokenNetwork, tokenAddress common.Address, token *contracts.Token, conn *ethclient.Client) {
	log.Printf("createchannel between %s-%s\n", utils.APex(account1), utils.APex(account2))
	auth1 := bind.NewKeyedTransactor(key1)
	auth2 := bind.NewKeyedTransactor(key2)

	//step 2.1 aprove
	approve := new(big.Int)
	approve = approve.Mul(amount, big.NewInt(100)) //保证多个通道创建的时候不会因为approve冲突
	tx, err := token.Approve(auth1, tokenNetworkAddress, approve)
	if err != nil {
		log.Fatalf("Failed to Approve: %v", err)
	}
	ctx := context.Background()
	_, err = bind.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to Approve when mining :%v", err)
	}
	fmt.Printf("Approve complete...\n")
	tx, err = token.Approve(auth2, tokenNetworkAddress, approve)
	if err != nil {
		log.Fatalf("Failed to Approve: %v", err)
	}
	ctx = context.Background()
	_, err = bind.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to Approve when mining :%v", err)
	}
	fmt.Printf("Approve complete...\n")
	// deposit
	tx, err = tokenNetwork.Deposit(auth1, tokenAddress, account1, account2, amount, 100)
	if err != nil {
		log.Printf("Failed to Deposit: %v,%s,%s", err, auth1.From.String(), account2.String())
		return
	}
	ctx = context.Background()
	_, err = bind.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to Deposit when mining :%v", err)
	}
	channelID, _, _, _, _, err := tokenNetwork.GetChannelInfo(nil, tokenAddress, account1, account2)
	log.Printf("create channel gas %s:%d,channel identifier=0x%s \n", tx.Hash().String(), tx.Gas(), hex.EncodeToString(channelID[:]))
	fmt.Printf("Deposit complete...\n")

	tx, err = tokenNetwork.Deposit(auth2, tokenAddress, account2, account1, amount, 100)
	if err != nil {
		log.Fatalf("Failed to Deposit2: %v", err)
	}
	ctx = context.Background()
	_, err = bind.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to Deposit when mining :%v", err)
	}
	fmt.Printf("Deposit complete...\n")
	time.Sleep(time.Millisecond * 10)
	return
}
