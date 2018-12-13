package zkp

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

//PublicParameters 公证人零知识证明参数
type PublicParameters struct {
	H1             *big.Int
	H2             *big.Int
	NTilde         *big.Int
	PaillierPubKey *PublicKey
}

type ECPoint struct {
	X *big.Int
	Y *big.Int
}

func (pp *PublicParameters) Initialization(curve *secp256k1.BitCurve,
	nTilde *big.Int,
	kPrime int32,
	h1, h2 *big.Int,
	paillierPubKey *PublicKey,
) {
	//pp.GRaw=curve.Encode()
	pp.NTilde = nTilde
	pp.H1 = h1
	pp.H2 = h2
	pp.PaillierPubKey = paillierPubKey
	return
}

/*func (pp PublicParameters)getG(curve *secp256k1.BitCurve)  ECPoint{
	//return curve.decodePoint(GRaw)
}*/
