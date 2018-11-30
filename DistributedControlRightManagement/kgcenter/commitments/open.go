package commitments

import (
	"math/big"

	"github.com/Nik-U/pbc"
)

type Open struct {
	secrets    []*big.Int
	randomness *pbc.Element
}

func (open *Open) Constructor(randomness *pbc.Element, secrets []*big.Int) {
	open.secrets = secrets
	open.randomness = randomness
}

func (open *Open) GetSecrets() []*big.Int {
	return open.secrets
}

func (open *Open) getRandomness() *pbc.Element {
	return open.randomness
}
