package commitments

import (
	"math/big"

	"github.com/Nik-U/pbc"
)

type MultiTrapdoorMasterPublicKey struct {
	g       *pbc.Element
	q       *big.Int
	h       *pbc.Element
	pairing *pbc.Pairing
}

func (mtmp *MultiTrapdoorMasterPublicKey) Constructor(g *pbc.Element,
	q *big.Int, h *pbc.Element, pairing *pbc.Pairing) {
	mtmp.g = g
	mtmp.q = q
	mtmp.h = h
	mtmp.pairing = pairing
}
