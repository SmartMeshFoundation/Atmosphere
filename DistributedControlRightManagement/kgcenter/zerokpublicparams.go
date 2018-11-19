package kgcenter

import (
	"math/big"
	//"github.com/Roasbeef/go-go-gadget-paillier"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type PublicParameters struct {
	gRaw           []byte
	h1             *big.Int
	h2             *big.Int
	nTilde         *big.Int
	paillierPubKey *PublicKey
}

type ECPoint struct {
	X *big.Int
	Y *big.Int
}

func (pp *PublicParameters)Initialization(curve *secp256k1.BitCurve,
	nTilde *big.Int,
	kPrime int32,
	h1,h2 *big.Int,
	paillierPubKey *PublicKey,
	) {
	//pp.gRaw=curve.Encode()
	pp.nTilde = nTilde
	pp.h1 = h1
	pp.h2 = h2
	pp.paillierPubKey = paillierPubKey
	return
}

/*func (pp PublicParameters)getG(curve *secp256k1.BitCurve)  ECPoint{
	//return curve.decodePoint(gRaw)
}*/