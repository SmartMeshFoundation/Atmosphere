package mecdsa

import (
	"crypto/rand"
	"testing"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/proofs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/share"
)

func TestMta(t *testing.T) {
	aliceInput := share.RandomPrivateKey()
	privAlice, err := proofs.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Error(err)
		return
	}
	bobInput := share.RandomPrivateKey()
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
	left := alpha.Clone()
	share.ModAdd(left, beta)
	right := aliceInput.Clone()
	share.ModMul(right, bobInput)
	if left.D.Cmp(right.D) != 0 {
		t.Error("not equal")
	}
}
