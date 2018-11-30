package kgcenter

/*import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/Nik-U/pbc"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/sirupsen/logrus"
)

type Commitment struct {
	committment *pbc.Element
	pubkey      *pbc.Element
}

//MultiTrapdoorCommitment
type MultiTrapdoorCommitment struct {
	commitment *Commitment
	open       *Open
}

type Open struct {
	secrets    []*big.Int
	randomness *pbc.Element
}

type CmtMasterPublicKey struct {
	g       *pbc.Element
	q       *big.Int
	h       *pbc.Element
	pairing *pbc.Pairing
}

func (ct *Commitment) New(pubkey *pbc.Element, a *pbc.Element) {
	ct.pubkey = pubkey
	ct.committment = a
}

func (open *Open) New(randomness *pbc.Element, secrets []*big.Int) {
	open.randomness = randomness

	open.secrets = secrets
}

func (open *Open) GetSecrets() []*big.Int {
	return open.secrets
}

func (open *Open) getRandomness() *pbc.Element {
	return open.randomness
}

func (mtdct *MultiTrapdoorCommitment) New(commitment *Commitment, open *Open) {
	mtdct.commitment = commitment
	mtdct.open = open
}

func (cmpk *CmtMasterPublicKey) New(g *pbc.Element, q *big.Int, h *pbc.Element, pairing *pbc.Pairing) {
	cmpk.g = g
	cmpk.q = q
	cmpk.h = h
	cmpk.pairing = pairing
}

func MultiLinnearCommit(rnd *rand.Rand, mpk *CmtMasterPublicKey, secrets []*big.Int) *MultiTrapdoorCommitment {
	e := mpk.pairing.NewZr()
	e.Rand()
	r := mpk.pairing.NewZr()
	r.Rand()

	h := func(target *pbc.Element, megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}

	secretsBytes := make([]string, len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen() + 7) / 8)
		se := make([]byte, count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}

	digest := mpk.pairing.NewZr()
	h(digest, secretsBytes[:])

	ge := mpk.pairing.NewG1()
	ge.MulZn(mpk.g, e)

	//he = mpk.h + ge
	he := mpk.pairing.NewG1()
	he.Add(mpk.h, ge)

	//he = r*he
	rhe := mpk.pairing.NewG1()
	rhe.MulZn(he, r)

	//dg = digest*mpk.g
	dg := mpk.pairing.NewG1()
	//
	dg.MulZn(mpk.g, digest)

	//a = mpk.g + he
	a := mpk.pairing.NewG1()
	a.Add(dg, rhe)

	open := new(Open)
	open.New(r, secrets)
	commitment := new(Commitment)
	commitment.New(e, a)

	mtdct := new(MultiTrapdoorCommitment)
	mtdct.New(commitment, open)

	return mtdct
}

//初始化配对pairing_init_set_str 0成功1失败
func GenerateMasterPK() *CmtMasterPublicKey {
	pairing, err := pbc.NewPairingFromString("type a\nq 461865797040674854785543520687465348175327491813790570884784113145245230030247605521999820430629387195211971356275124424683576752355352508847321749377791\nh 632042503420655948751805161979977007676946285961712997135601731248604852746868244763371913442410154700032\nr 730751167114595186142829002853739519958614802431\nexp2 159\nexp1 138\nsign1 1\nsign0 -1\n")
	if err != nil {
		fmt.Println("preload pairing fail.\n")
	}

	g := getBasePoint(pairing)
	q, _ := new(big.Int).SetString("730751167114595186142829002853739519958614802431", 10)
	h := RandomPointInG1(pairing)
	cmpk := new(CmtMasterPublicKey)
	cmpk.New(g, q, h, pairing)
	return cmpk
}

func RandomPointInG1(pairing *pbc.Pairing) *pbc.Element {
	for {
		h := pairing.NewG1()
		h.Rand()

		cof := pairing.NewZr()
		num, _ := new(big.Int).SetString("632042503420655948751805161979977007676946285961712997135601731248604852746868244763371913442410154700032", 10)
		cof.SetBig(num)

		hh := pairing.NewG1()
		hh.MulZn(h, cof)

		order, _ := new(big.Int).SetString("730751167114595186142829002853739519958614802431", 10)
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
	num, _ := new(big.Int).SetString("632042503420655948751805161979977007676946285961712997135601731248604852746868244763371913442410154700032", 10)
	cof.SetBig(num)

	order, _ := new(big.Int).SetString("730751167114595186142829002853739519958614802431", 10)
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

func (mtc *MultiTrapdoorCommitment) CmtOpen() *Open {
	return mtc.open
}

func (mtc *MultiTrapdoorCommitment) CmtCommitment() *Commitment {
	return mtc.commitment
}

//
func Checkcommitment(commitment *Commitment, open *Open, mpk *CmtMasterPublicKey) bool {
	g := mpk.g
	h := mpk.h

	f := func(target *pbc.Element, megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}

	secrets := open.GetSecrets()
	secretsBytes := make([]string, len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen() + 7) / 8)
		se := make([]byte, count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}

	digest := mpk.pairing.NewZr()
	f(digest, secretsBytes[:])

	rg := mpk.pairing.NewG1()
	rg.MulZn(g, open.getRandomness())

	d1 := mpk.pairing.NewG1()
	d1.MulZn(g, commitment.pubkey)

	dh := mpk.pairing.NewG1()
	dh.Add(h, d1)

	gdn := mpk.pairing.NewG1()
	digest.Neg(digest)
	gdn.MulZn(g, digest)

	comd := mpk.pairing.NewG1()
	comd.Add(commitment.committment, gdn)
	b := DDHTest(rg, dh, comd, g, mpk.pairing)
	if b == false {
		logrus.Error("Check commitment error")
	}
	return b
}

//见 main2 pairing(g,h)^(a*b)
func DDHTest(a *pbc.Element, b *pbc.Element, c *pbc.Element, generator *pbc.Element, pairing *pbc.Pairing) bool {

	temp1 := pairing.NewGT().Pair(a, b)
	temp2 := pairing.NewGT().Pair(generator, c)

	return temp1.Equals(temp2) //temp1=temp2
} */
