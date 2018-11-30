package paillier

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"time"
)

type ThresholdKeyGenerator struct {
	PublicKeyBitLength             int
	TotalNumberOfDecryptionServers int
	Threshold                      int
	random                         io.Reader

	p *big.Int // p is prime of `PublicKeyBitLength/2` bits and `p = 2*p1 + 1`
	q *big.Int // q is prime of `PublicKeyBitLength/2` bits and `q = 2*q1 + 1`

	p1 *big.Int // p1 is prime of `PublicKeyBitLength/2 - 1` bits
	q1 *big.Int // q1 is prime of `PublicKeyBitLength/2 - 1` bits

	n       *big.Int // n=p*q and is of `PublicKeyBitLength` bits
	m       *big.Int // m = p1*q1
	nSquare *big.Int // nSquare = n*n
	nm      *big.Int // nm = n*m

	// As specified in the paper, d must satify d=1 mod n and d=0 mod m
	d *big.Int

	// A generator of QR in Z_{n^2}
	v *big.Int

	// The polynomial coefficients to hide a secret. See Shamir.
	polynomialCoefficients []*big.Int
}

func GetThresholdKeyGenerator(
	publicKeyBitLength int,
	totalNumberOfDecryptionServers int,
	threshold int,
	random io.Reader,
) (*ThresholdKeyGenerator, error) {
	if publicKeyBitLength%2 == 1 {
		return nil, errors.New("Public key bit length must be an even number")
	}
	if publicKeyBitLength < 18 {
		return nil, errors.New("Public key bit length must be at least 18 bits")
	}

	return &ThresholdKeyGenerator{
		PublicKeyBitLength:             publicKeyBitLength,
		TotalNumberOfDecryptionServers: totalNumberOfDecryptionServers,
		Threshold:                      threshold,
		random:                         random,
	}, nil
}

func (tkg *ThresholdKeyGenerator) generateSafePrimes() (*big.Int, *big.Int, error) {
	concurrencyLevel := 4
	timeout := 120 * time.Second
	safePrimeBitLength := tkg.PublicKeyBitLength / 2

	return GenerateSafePrime(safePrimeBitLength, concurrencyLevel, timeout, tkg.random)
}

func (tkg *ThresholdKeyGenerator) initPandP1() error {
	var err error
	tkg.p, tkg.p1, err = tkg.generateSafePrimes()
	return err
}

func (tkg *ThresholdKeyGenerator) initQandQ1() error {
	var err error
	tkg.q, tkg.q1, err = tkg.generateSafePrimes()
	return err
}

func (tkg *ThresholdKeyGenerator) initShortcuts() {
	tkg.n = new(big.Int).Mul(tkg.p, tkg.q)
	tkg.m = new(big.Int).Mul(tkg.p1, tkg.q1)
	tkg.nSquare = new(big.Int).Mul(tkg.n, tkg.n)
	tkg.nm = new(big.Int).Mul(tkg.n, tkg.m)
}

func (tkg *ThresholdKeyGenerator) arePsAndQsGood() bool {
	if tkg.p.Cmp(tkg.q) == 0 {
		return false
	}
	if tkg.p.Cmp(tkg.q1) == 0 {
		return false
	}
	if tkg.p1.Cmp(tkg.q) == 0 {
		return false
	}
	return true
}

func (tkg *ThresholdKeyGenerator) initPsAndQs() error {
	if err := tkg.initPandP1(); err != nil {
		return err
	}
	if err := tkg.initQandQ1(); err != nil {
		return err
	}
	if !tkg.arePsAndQsGood() {
		return tkg.initPsAndQs()
	}
	return nil
}

func (tkg *ThresholdKeyGenerator) computeV() error {
	var err error
	tkg.v, err = GetRandomGeneratorOfTheQuadraticResidue(tkg.nSquare, tkg.random)
	return err
}

func (tkg *ThresholdKeyGenerator) initD() {
	mInverse := new(big.Int).ModInverse(tkg.m, tkg.n)
	tkg.d = new(big.Int).Mul(mInverse, tkg.m)
}

func (tkg *ThresholdKeyGenerator) initNumerialValues() error {
	if err := tkg.initPsAndQs(); err != nil {
		return err
	}
	tkg.initShortcuts()
	tkg.initD()
	return tkg.computeV()
}

func (tkg *ThresholdKeyGenerator) generateHidingPolynomial() error {
	tkg.polynomialCoefficients = make([]*big.Int, tkg.Threshold)
	tkg.polynomialCoefficients[0] = tkg.d
	var err error
	for i := 1; i < tkg.Threshold; i++ {
		tkg.polynomialCoefficients[i], err = rand.Int(tkg.random, tkg.nm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tkg *ThresholdKeyGenerator) computeShare(index int) *big.Int {
	share := big.NewInt(0)
	for i := 0; i < tkg.Threshold; i++ {
		a := tkg.polynomialCoefficients[i]
		b := new(big.Int).Exp(big.NewInt(int64(index+1)), big.NewInt(int64(i)), nil)
		tmp := new(big.Int).Mul(a, b)
		share = new(big.Int).Add(share, tmp)
	}
	return new(big.Int).Mod(share, tkg.nm)
}

func (tkg *ThresholdKeyGenerator) createShares() []*big.Int {
	shares := make([]*big.Int, tkg.TotalNumberOfDecryptionServers)
	for i := 0; i < tkg.TotalNumberOfDecryptionServers; i++ {
		shares[i] = tkg.computeShare(i)
	}
	return shares
}

func (tkg *ThresholdKeyGenerator) delta() *big.Int {
	return Factorial(tkg.TotalNumberOfDecryptionServers)
}

func (tkg *ThresholdKeyGenerator) createViArray(shares []*big.Int) (viArray []*big.Int) {
	viArray = make([]*big.Int, len(shares))
	delta := tkg.delta()
	for i, share := range shares {
		tmp := new(big.Int).Mul(share, delta)
		viArray[i] = new(big.Int).Exp(tkg.v, tmp, tkg.nSquare)
	}
	return viArray
}

func (tkg *ThresholdKeyGenerator) createPrivateKey(i int, share *big.Int, viArray []*big.Int) *ThresholdPrivateKey {
	ret := new(ThresholdPrivateKey)
	ret.N = tkg.n
	ret.V = tkg.v

	ret.TotalNumberOfDecryptionServers = tkg.TotalNumberOfDecryptionServers
	ret.Threshold = tkg.Threshold
	ret.Share = share
	ret.Id = i + 1
	ret.Vi = viArray
	return ret
}

func (tkg *ThresholdKeyGenerator) createPrivateKeys() []*ThresholdPrivateKey {
	shares := tkg.createShares()
	viArray := tkg.createViArray(shares)
	ret := make([]*ThresholdPrivateKey, tkg.TotalNumberOfDecryptionServers)
	for i := 0; i < tkg.TotalNumberOfDecryptionServers; i++ {
		ret[i] = tkg.createPrivateKey(i, shares[i], viArray)
	}
	return ret
}

func (tkg *ThresholdKeyGenerator) Generate() ([]*ThresholdPrivateKey, error) {
	if err := tkg.initNumerialValues(); err != nil {
		return nil, err
	}
	if err := tkg.generateHidingPolynomial(); err != nil {
		return nil, err
	}
	return tkg.createPrivateKeys(), nil
}
