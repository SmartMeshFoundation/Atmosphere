package models

import (
	"math/big"
	"testing"

	"encoding/json"
	"reflect"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
)

func TestNewPrivateKeyInfo(t *testing.T) {
	db := SetupTestDB()
	defer db.Close()
	r := utils.NewRandomHash()
	p := &PrivateKeyInfo{
		Key: r,
		K:   big.NewInt(30),
		Zkps: []*Zkp{
			{NotaryName: "n1", EncryptedK: big.NewInt(30)},
			{NotaryName: "n2", EncryptedK: big.NewInt(40)},
		},
	}
	err := db.NewPrivateKeyInfo(p)
	if err != nil {
		t.Error(err)
	}
	p2, err := db.LoadPrivatedKeyInfo(p.Key)
	if err != nil {
		t.Error(err)
		return
	}
	if len(p2.Zkps) != 2 {
		t.Error("there should be 2 zpks")
		return
	}
	t.Logf("p=%s", utils.StringInterface(p, 5))
	t.Logf("p2=%s", utils.StringInterface(p2, 5))
	err = db.AddZkp(&Zkp{
		Key:        p.Key,
		NotaryName: "n3",
	})
	if err != nil {
		t.Error(err)
		return
	}
	p2, err = db.LoadPrivatedKeyInfo(p.Key)
	if err != nil {
		t.Error(err)
		return
	}
	if len(p2.Zkps) != 3 {
		t.Error("there should be 3 zpks")
		return
	}
}

func TestJsonZkp(t *testing.T) {
	z := new(Zkp)
	z.Key = utils.NewRandomHash()
	z.Zkp = &zkp.Zkp{
		Z: big.NewInt(20),
		E: big.NewInt(30),
	}
	z.PublicKeyX = big.NewInt(2)
	z.PublicKeyY = big.NewInt(7)
	z.EncryptedK = big.NewInt(1)
	data, err := json.Marshal(z)
	if err != nil {
		t.Error(err)
		return
	}
	z2 := new(Zkp)
	err = json.Unmarshal(data, z2)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(z, z2) {
		t.Error("not equal")
		return
	}
}
