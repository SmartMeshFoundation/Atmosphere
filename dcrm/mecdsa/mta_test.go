package mecdsa

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/proofs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/secret_sharing"
)

func TestMta(t *testing.T) {
	aliceInput := secret_sharing.RandomPrivateKey()
	privAlice, err := proofs.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Error(err)
		return
	}
	bobInput := secret_sharing.RandomPrivateKey()
	ma, err := NewMessageA(aliceInput, &privAlice.PublicKey)
	if err != nil {
		t.Error(err)
		return
	}
	mb, beta, err := NewMessageB(bobInput, &privAlice.PublicKey, ma)
	alpha, err := mb.VerifyProofsGetAlpha(privAlice, aliceInput)
	if err != nil {
		t.Error(err)
		return
	}
	left := new(big.Int).Set(alpha)
	secret_sharing.ModAdd(left, beta)
	right := new(big.Int).Set(aliceInput)
	secret_sharing.ModMul(right, bobInput)
	if left.Cmp(right) != 0 {
		t.Error("not equal")
	}
}
