package models

import (
	"math/big"

	"encoding/json"
	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/ethereum/go-ethereum/common"
)

//对一个具体的Tx进行签名,使用PrivateKey对应的Key
type Tx struct {
	Key        common.Hash //应该是TxHash
	PrivateKey common.Hash
	Message    []byte //待签数据

	Zkpi1s []*Zkpi1
	Zkpi2s []*Zkpi2
}

//用总私钥进行一次签名
type TxModel struct {
	Key        []byte `gorm:"primary_key"` //应该是TxHash
	PrivateKey []byte `gorm:"index"`
	Message    []byte //待签数据

	Zkpi1s []byte `gorm:"size:4096"`
	Zkpi2s []byte `gorm:"size:4096"`
}
type Zkpi1CommitmentArg struct {
	RhoI    *big.Int
	RhoIRnd *big.Int
	UI      *big.Int
	VI      *big.Int
}
type Zkpi1 struct {
	Zkpi1      *zkp.Zkpi1
	Arg        *Zkpi1CommitmentArg
	Cmt        *commitments.MultiTrapdoorCommitment
	NotaryName string //which notary
}
type Zkpi2CommitmentArg struct {
	KI    *big.Int
	CI    *big.Int
	CIRnd *big.Int
	RIx   *big.Int
	RIy   *big.Int
	Mask  *big.Int
	WI    *big.Int
}
type Zkpi2 struct {
	Zkpi2      *zkp.Zkpi2
	Arg        *Zkpi2CommitmentArg
	Cmt        *commitments.MultiTrapdoorCommitment
	NotaryName string //which notary
}

func fromTxModel(t *TxModel) *Tx {
	t2 := &Tx{
		Message: t.Message,
	}
	t2.Key.SetBytes(t.Key)
	t2.PrivateKey.SetBytes(t.PrivateKey)
	if len(t.Zkpi1s) > 10 {
		err := json.Unmarshal(t.Zkpi1s, &t2.Zkpi1s)
		if err != nil {
			panic(fmt.Sprintf("Unmarshal Zkpi1s err %s", err))
		}
	}
	if len(t.Zkpi2s) > 10 {
		err := json.Unmarshal(t.Zkpi2s, &t2.Zkpi2s)
		if err != nil {
			panic(fmt.Sprintf("Unmarshal Zkpi2s err %s", err))
		}
	}
	return t2
}

func toTxModel(t *Tx) *TxModel {
	t2 := &TxModel{
		Key:        t.Key[:],
		PrivateKey: t.PrivateKey[:],
		Message:    t.Message,
	}
	if len(t.Zkpi1s) > 0 {
		var err error
		t2.Zkpi1s, err = json.Marshal(t.Zkpi1s)
		if err != nil {
			panic(fmt.Sprintf("marshal Zkpi1s err %s", err))
		}
	}
	if len(t.Zkpi2s) > 0 {
		var err error
		t2.Zkpi2s, err = json.Marshal(t.Zkpi2s)
		if err != nil {
			panic(fmt.Sprintf("marshal Zkpi2s err %s", err))
		}
	}
	return t2
}

func (db *DB) NewTx(t *Tx) error {
	return db.Create(toTxModel(t)).Error
}

func (db *DB) LoadTx(key common.Hash) (*Tx, error) {
	var txModel TxModel
	err := db.Where(&TxModel{
		Key: key[:],
	}).Find(&txModel).Error
	if err != nil {
		return nil, err
	}
	return fromTxModel(&txModel), nil
}

func (db *DB) UpdateTxZkpi1(t *Tx) error {
	data, err := json.Marshal(t.Zkpi1s)
	if err != nil {
		return err
	}
	return db.Model(&TxModel{
		Key: t.Key[:],
	}).Update(&TxModel{
		Zkpi1s: data,
	}).Error
}

func (db *DB) UpdateTxZkpi2(t *Tx) error {
	data, err := json.Marshal(t.Zkpi2s)
	if err != nil {
		return err
	}
	return db.Model(&TxModel{
		Key: t.Key[:],
	}).Update(&TxModel{
		Zkpi2s: data,
	}).Error
}
