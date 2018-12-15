package mecdsa

import (
	"math/big"

	"errors"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/proofs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/share"
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
func NewMessageA(ecdsaPrivateKey share.SPrivKey, paillierPubKey *proofs.PublicKey) (*MessageA, error) {
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
func NewMessageB(ecdsaPrivateKey share.SPrivKey, paillierPubKey *proofs.PublicKey, ca *MessageA) (*MessageB, share.SPrivKey, error) {
	betaTag := zkp.RandomFromZn(paillierPubKey.N)
	//todo fixme bai
	//betaTag = big.NewInt(39)
	betaTagPrivateKey := share.BigInt2PrivateKey(betaTag.Mod(betaTag, share.S.N))
	cBetaTag, err := proofs.Encrypt(paillierPubKey, betaTagPrivateKey.Bytes())
	if err != nil {
		return nil, share.SPrivKey{}, err
	}
	bca := proofs.Mul(paillierPubKey, ca.C, ecdsaPrivateKey.Bytes())
	cb := proofs.AddCipher(paillierPubKey, bca, cBetaTag)
	beta := share.ModSub(share.PrivKeyZero.Clone(), betaTagPrivateKey)
	bproof := proofs.Prove(ecdsaPrivateKey)
	betaTagProof := proofs.Prove(betaTagPrivateKey)
	return &MessageB{
		C:            cb,
		BProof:       bproof,
		BetaTagProof: betaTagProof,
	}, beta, nil
}

//const
func (m *MessageB) VerifyProofsGetAlpha(dk *proofs.PrivateKey, a share.SPrivKey) (share.SPrivKey, error) {
	ashare, err := proofs.Decrypt(dk, m.C)
	if err != nil {
		return share.SPrivKey{}, err
	}
	alpha := new(big.Int).SetBytes(ashare)
	alphaKey := share.BigInt2PrivateKey(alpha)
	gAlphaX, gAlphaY := share.S.ScalarBaseMult(alphaKey.Bytes())
	babTagX, babTagY := share.S.ScalarMult(m.BProof.PK.X, m.BProof.PK.Y, a.Bytes())
	babTagX, babTagY = share.PointAdd(babTagX, babTagY, m.BetaTagProof.PK.X, m.BetaTagProof.PK.Y)
	if proofs.Verify(m.BProof) && proofs.Verify(m.BetaTagProof) &&
		babTagX.Cmp(gAlphaX) == 0 &&
		babTagY.Cmp(gAlphaY) == 0 {
		return alphaKey, nil
	}
	return share.SPrivKey{}, errors.New("invalid key")
}

func EqualGE(pubGB *share.SPubKey, mtaGB *share.SPubKey) bool {
	return pubGB.X.Cmp(mtaGB.X) == 0 && pubGB.Y.Cmp(mtaGB.Y) == 0
}
