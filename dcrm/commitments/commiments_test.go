package commitments

import (
	"testing"

	"encoding/json"
	"math/big"
	"reflect"
)

func TestGenerateNMMasterPublicKey(t *testing.T) {
	k := GenerateNMMasterPublicKey()
	data, err := json.Marshal(k)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("data=%s", string(data))
	k2 := &MultiTrapdoorMasterPublicKey{}
	err = json.Unmarshal(data, k2)
	if err != nil {
		t.Error(err)
		return
	}
	if k.Q.Cmp(k2.Q) != 0 {
		t.Errorf("not equal")
	}
}

func TestOpen_MarshalJSON(t *testing.T) {
	open := new(Open)
	r := DefaultPairing.NewZr()
	secrets := []*big.Int{big.NewInt(1111111), big.NewInt(2)}
	open.Constructor(r, secrets)
	data, err := json.Marshal(open)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("data=%s", string(data))
	open2 := new(Open)
	err = json.Unmarshal(data, open2)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(secrets, open2.secrets) {
		t.Error("not equal")
	}

	open.Constructor(r, nil)
	data, err = json.Marshal(open)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("data=%s", string(data))
	open2 = new(Open)
	err = json.Unmarshal(data, open2)
	if err != nil {
		t.Error(err)
		return
	}
	if open2.secrets != nil {
		t.Error("not equal")
	}

}

func TestCommitment(t *testing.T) {
	commitment := new(Commitment)
	commitment.Constructor(DefaultPairing.NewZr(), DefaultPairing.NewG1())
	data, err := json.Marshal(commitment)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("data=%s", string(data))
	commitment2 := new(Commitment)
	err = json.Unmarshal(data, commitment2)
	if err != nil {
		t.Error(err)
		return
	}
}
func TestElementJson(t *testing.T) {
	k := GenerateNMMasterPublicKey()
	data, err := json.Marshal(k.G)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("data=%s", string(data))
	err = json.Unmarshal(data, k.H)
	if err != nil {
		t.Error(err)
		return
	}
}
