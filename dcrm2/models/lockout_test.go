package models

import (
	"testing"

	"math/big"

	"reflect"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
)

func TestNewTx(t *testing.T) {
	db := SetupTestDB()
	defer db.Close()
	tx := &Tx{
		Key:        utils.NewRandomHash(),
		PrivateKey: utils.NewRandomHash(),
		Message:    []byte("aaa"),
	}
	tx.Zkpi1s = []*Zkpi1{
		{
			Zkpi1: &zkp.Zkpi1{
				Z:  big.NewInt(20),
				U1: big.NewInt(2),
			},
			Arg: &Zkpi1CommitmentArg{
				RhoI:    big.NewInt(3),
				RhoIRnd: big.NewInt(4),
			},
		},
	}
	tx.Zkpi2s = []*Zkpi2{
		{
			Zkpi2: &zkp.Zkpi2{
				Z1: big.NewInt(4),
				U2: big.NewInt(5),
			},
			Arg: &Zkpi2CommitmentArg{
				KI: big.NewInt(6),
				CI: big.NewInt(7),
			},
		},
	}
	err := db.NewTx(tx)
	if err != nil {
		t.Error(err)
		return
	}
	tx2, err := db.LoadTx(tx.Key)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(tx2, tx) {
		t.Error("not equal")
		return
	}
	tx.Zkpi1s = append(tx.Zkpi1s, &Zkpi1{
		Zkpi1: &zkp.Zkpi1{
			Z:  big.NewInt(20),
			U1: big.NewInt(2),
		},
		Arg: &Zkpi1CommitmentArg{
			RhoI:    big.NewInt(3),
			RhoIRnd: big.NewInt(4),
		},
	})
	tx.Zkpi2s = append(tx.Zkpi2s, &Zkpi2{
		Zkpi2: &zkp.Zkpi2{
			Z1: big.NewInt(4),
			U2: big.NewInt(5),
		},
		Arg: &Zkpi2CommitmentArg{
			KI: big.NewInt(6),
			CI: big.NewInt(7),
		},
	})
	err = db.UpdateTxZkpi1(tx)
	if err != nil {
		t.Error(err)
		return
	}
	tx2, err = db.LoadTx(tx.Key)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(tx.Zkpi1s, tx.Zkpi1s) {
		t.Error("must equal")
		return
	}
	if reflect.DeepEqual(tx2.Zkpi2s, tx.Zkpi2s) {
		t.Error("must not equal")
		return
	}
	err = db.UpdateTxZkpi2(tx)
	if err != nil {
		t.Error(err)
	}
	tx2, err = db.LoadTx(tx.Key)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(tx2, tx) {
		t.Error("must equal")
		return
	}
}
