package mecdsa

import (
	"math/big"

	"errors"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/proofs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/secret_sharing"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm2/zkp"
)

type MessageA struct {
	C []byte //paillier encion 文本
}
type MessageB struct {
	C            []byte //pailler加密文本
	BProof       *proofs.DLogProof
	BetaTagProof *proofs.DLogProof
}

//const
func NewMessageA(ecdsaPrivateKey *big.Int, paillierPubKey *proofs.PublicKey) (*MessageA, error) {
	ca, err := proofs.Encrypt(paillierPubKey, ecdsaPrivateKey.Bytes())
	if err != nil {
		return nil, err
	}
	return &MessageA{ca}, nil
}
func (m *MessageA) String() string {
	return new(big.Int).SetBytes(m.C).Text(10)
}

//const
func NewMessageB(ecdsaPrivateKey *big.Int, paillierPubKey *proofs.PublicKey, ca *MessageA) (*MessageB, *big.Int, error) {
	betaTag := zkp.RandomFromZn(paillierPubKey.N)
	//todo fixme bai
	//betaTag = big.NewInt(39)
	betaTagPrivateKey := betaTag.Mod(betaTag, secret_sharing.S.N)
	cBetaTag, err := proofs.Encrypt(paillierPubKey, betaTagPrivateKey.Bytes())
	if err != nil {
		return nil, nil, err
	}
	bca := proofs.Mul(paillierPubKey, ca.C, ecdsaPrivateKey.Bytes())
	cb := proofs.AddCipher(paillierPubKey, bca, cBetaTag)
	beta := secret_sharing.ModSub(big.NewInt(0), betaTagPrivateKey)
	bproof := proofs.Prove(ecdsaPrivateKey)
	betaTagProof := proofs.Prove(betaTagPrivateKey)
	return &MessageB{
		C:            cb,
		BProof:       bproof,
		BetaTagProof: betaTagProof,
	}, beta, nil
}

//const
func (m *MessageB) VerifyProofsGetAlpha(dk *proofs.PrivateKey, a *big.Int) (*big.Int, error) {
	ashare, err := proofs.Decrypt(dk, m.C)
	if err != nil {
		return nil, err
	}
	alpha := new(big.Int).SetBytes(ashare)
	alpha.Mod(alpha, secret_sharing.S.N)
	gAlphaX, gAlphaY := secret_sharing.S.ScalarBaseMult(alpha.Bytes())
	babTagX, babTagY := secret_sharing.S.ScalarMult(m.BProof.PK.X, m.BProof.PK.Y, a.Bytes())
	babTagX, babTagY = secret_sharing.PointAdd(babTagX, babTagY, m.BetaTagProof.PK.X, m.BetaTagProof.PK.Y)
	if proofs.Verify(m.BProof) && proofs.Verify(m.BetaTagProof) &&
		babTagX.Cmp(gAlphaX) == 0 &&
		babTagY.Cmp(gAlphaY) == 0 {
		return alpha, nil
	}
	return nil, errors.New("invalid key")
}

func EqualGE(pubGB *secret_sharing.GE, mtaGB *secret_sharing.GE) bool {
	return pubGB.X.Cmp(mtaGB.X) == 0 && pubGB.Y.Cmp(mtaGB.Y) == 0
}
