package commitments

import "github.com/Nik-U/pbc"

type Commitment struct {
	pubkey      *pbc.Element
	committment *pbc.Element
}

func (c *Commitment) Constructor(pubkey *pbc.Element, a *pbc.Element) {
	c.pubkey = pubkey
	c.committment = a
}
