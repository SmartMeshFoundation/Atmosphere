package commitments

import (
	"fmt"
	"math/big"

	"github.com/Nik-U/pbc"
)

const (
	qOfPBC     = "461865797040674854785543520687465348175327491813790570884784113145245230030247605521999820430629387195211971356275124424683576752355352508847321749377791"
	hOfPBC     = "632042503420655948751805161979977007676946285961712997135601731248604852746868244763371913442410154700032"
	rOfPBC     = "730751167114595186142829002853739519958614802431"
	exp2OfPBC  = "159"
	exp1OfPBC  = "138"
	sign1OfPBC = "1"
	sign0OfPBC = "-1"
)

const ConsensusParamOfPBC = "type a\n" +
	"q " + qOfPBC + "\n" +
	"h " + hOfPBC + "\n" +
	"r " + rOfPBC + "\n" +
	"exp2 " + exp2OfPBC + "\n" +
	"exp1 " + exp1OfPBC + "\n" +
	"sign1 " + sign1OfPBC + "\n" +
	"sign0 " + sign0OfPBC + "\n"

func GenerateNMMasterPublicKey() *MultiTrapdoorMasterPublicKey {
	pairing, err := pbc.NewPairingFromString(ConsensusParamOfPBC)
	if err != nil {
		fmt.Println("preload pairing fail.\n")
	}
	g := getBasePoint(pairing)
	q, _ := new(big.Int).SetString(rOfPBC, 10)
	h := randomPointInG1(pairing)
	cmpk := new(MultiTrapdoorMasterPublicKey)
	cmpk.Constructor(g, q, h, pairing)
	return cmpk
}

func randomPointInG1(pairing *pbc.Pairing) *pbc.Element {
	for {
		h := pairing.NewG1()
		h.Rand()

		cof := pairing.NewZr()
		num, _ := new(big.Int).SetString(hOfPBC, 10)
		cof.SetBig(num)

		hh := pairing.NewG1()
		hh.MulZn(h, cof)

		order, _ := new(big.Int).SetString(rOfPBC, 10)
		q := pairing.NewZr()
		q.SetBig(order)

		hhh := pairing.NewG1()
		hhh.MulZn(hh, q)

		if hhh.Is0() {
			return hh
		}
	}
	return nil
}

func getBasePoint(pairing *pbc.Pairing) *pbc.Element {
	var p *pbc.Element
	cof := pairing.NewZr()
	num, _ := new(big.Int).SetString(hOfPBC, 10)
	cof.SetBig(num)

	order, _ := new(big.Int).SetString(rOfPBC, 10)
	q := pairing.NewZr()
	q.SetBig(order)

	for {
		p = pairing.NewG1()
		p.Rand()
		ge := pairing.NewG1()
		ge.MulZn(p, cof)

		pq := pairing.NewG1()
		pq.MulZn(ge, q)

		if ge.Is0() || pq.Is0() {
			return ge
		}
	}
	return nil
}
