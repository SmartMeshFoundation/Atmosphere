package paillier

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/ncw/gmp"
)

type Ciphertext struct {
	C *big.Int
}

type PublicKey struct {
	N          *big.Int //N=p*q
	G          *big.Int // usually G is set to N+1
	N2         *big.Int // the cache value of N^2
	K          int      // message space 2^K < N
	S          int      // security parameter for statistical secure MPC
	P          *big.Int // secret share prime
	FPPrecBits int      // fixed point precision bits
}

type SecretKey struct {
	PublicKey
	Lambda, Lm, Mu *big.Int
}

func (pk *PublicKey) GetNSquare() *big.Int {
	if pk.N2 != nil {
		return pk.N2
	}
	pk.N2 = new(big.Int).Mul(pk.N, pk.N)
	return pk.N2
}

func (pk *PublicKey) EAdd(cts ...*Ciphertext) *Ciphertext {
	accumulator := big.NewInt(1)

	for _, c := range cts {
		accumulator = new(big.Int).Mod(
			new(big.Int).Mul(accumulator, c.C),
			pk.GetNSquare(),
		)
	}

	return &Ciphertext{
		C: accumulator,
	}
}

func (pk *PublicKey) ESub(cts ...*Ciphertext) *Ciphertext {

	accumulator := cts[0].C

	for i, c := range cts {
		if i == 0 {
			continue
		}
		neg := new(big.Int).ModInverse(c.C, pk.GetNSquare())
		accumulator = new(big.Int).Mod(
			new(big.Int).Mul(accumulator, neg),
			pk.GetNSquare(),
		)
	}

	return &Ciphertext{
		C: accumulator,
	}
}

func (pk *PublicKey) ECMult(ct *Ciphertext, k *big.Int) *Ciphertext {

	gmpC := gmp.NewInt(0).SetBytes(ct.C.Bytes())
	gmpK := gmp.NewInt(0).SetBytes(k.Bytes())
	gmpN2 := gmp.NewInt(0).SetBytes(pk.GetNSquare().Bytes())

	m := new(gmp.Int).Exp(gmpC, gmpK, gmpN2)
	return &Ciphertext{new(big.Int).SetBytes(m.Bytes())}
}

func (sk *SecretKey) String() string {
	ret := fmt.Sprintf("g     :  %s\n", sk.G.String())
	ret += fmt.Sprintf("n     :  %s\n", sk.N.String())
	ret += fmt.Sprintf("lambda:  %s\n", sk.Lambda.String())
	ret += fmt.Sprintf("mu    :  %s\n", sk.Mu.String())
	return ret
}

func (sk *SecretKey) Decrypt(ciphertext *Ciphertext) *big.Int {

	gmpLambda := gmp.NewInt(0).SetBytes(sk.Lambda.Bytes())
	gmpC := gmp.NewInt(0).SetBytes(ciphertext.C.Bytes())
	gmpN2 := gmp.NewInt(0).SetBytes(sk.GetNSquare().Bytes())
	gmpTmp := new(gmp.Int).Exp(gmpC, gmpLambda, gmpN2)

	tmp := new(big.Int).SetBytes(gmpTmp.Bytes())
	mu := new(big.Int).ModInverse(sk.Lambda, sk.N)
	m := new(big.Int).Mod(new(big.Int).Mul(L(tmp, sk.N), mu), sk.N)
	return m
}

func (pk *PublicKey) EncryptWithR(m *big.Int, r *big.Int) *Ciphertext {

	nSquare := pk.GetNSquare()

	g := new(big.Int).Add(pk.N, big.NewInt(1))
	gmpG := gmp.NewInt(0).SetBytes(g.Bytes())
	gmpM := gmp.NewInt(0).SetBytes(m.Bytes())
	gmpN := gmp.NewInt(0).SetBytes(pk.N.Bytes())
	gmpN2 := gmp.NewInt(0).SetBytes(nSquare.Bytes())
	gmpR := gmp.NewInt(0).SetBytes(r.Bytes())
	gm := new(gmp.Int).Exp(gmpG, gmpM, gmpN2)
	rn := new(gmp.Int).Exp(gmpR, gmpN, gmpN2)
	c := new(gmp.Int).Mod(new(gmp.Int).Mul(rn, gm), gmpN2)
	return &Ciphertext{new(big.Int).SetBytes(c.Bytes())}
}

func (pk *PublicKey) Encrypt(m *big.Int) *Ciphertext {

	var r *big.Int
	var err error
	for {
		r, err = GetRandomNumberInMultiplicativeGroup(pk.N, rand.Reader)
		if err == nil {
			break
		}
	}
	return pk.EncryptWithR(m, r)
}

func L(u, n *big.Int) *big.Int {
	t := new(big.Int).Sub(u, big.NewInt(1))
	return new(big.Int).Div(t, n)
}

func LCM(x, y *big.Int) *big.Int {
	return new(big.Int).Mul(new(big.Int).Div(x, new(big.Int).GCD(nil, nil, x, y)), y)
}

func minusOne(x *big.Int) *big.Int {
	return new(big.Int).Add(x, big.NewInt(-1))
}

func computeMu(g, lambda, n *big.Int) *big.Int {
	n2 := new(big.Int).Mul(n, n)
	u := new(big.Int).Exp(g, lambda, n2)
	return new(big.Int).ModInverse(L(u, n), n)
}

func computePhi(p, q *big.Int) *big.Int {
	return new(big.Int).Mul(minusOne(p), minusOne(q))
}

func CreateKeyPair(bits int) (*SecretKey, *PublicKey) {

	// generate the prime factors
	var p *big.Int
	var q *big.Int
	var err error
	for {
		p, err = rand.Prime(rand.Reader, bits)
		if err != nil {
			continue
		}
		q, err = rand.Prime(rand.Reader, bits)
		if err != nil {
			continue
		}

		if p.Cmp(q) == 0 {
			continue
		}
		break
	}

	n := new(big.Int).Mul(p, q)
	lambda := computePhi(p, q)

	pk := &PublicKey{
		N: n,
	}

	return &SecretKey{
		PublicKey: *pk,
		Lambda:    lambda,
	}, pk
}
