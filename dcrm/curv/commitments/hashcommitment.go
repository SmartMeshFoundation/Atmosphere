package commitments

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

func CreateCommitmentWithUserDefinedRandomNess(message *big.Int, blindingFactor *big.Int) *big.Int {
	hash := crypto.Keccak256Hash(message.Bytes(), blindingFactor.Bytes())
	b := new(big.Int)
	b.SetBytes(hash[:])
	return b
}
