package paillier

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"

	"github.com/ncw/gmp"
)

type ThresholdPublicKey struct {
	PublicKey
	TotalNumberOfDecryptionServers int
	Threshold                      int
	V                              *big.Int   // needed for ZKP
	Vi                             []*big.Int // needed for ZKP
}

func (tk *ThresholdPublicKey) combineSharesConstant() *big.Int {
	tmp := new(big.Int).Mul(FOUR, new(big.Int).Mul(tk.delta(), tk.delta()))
	return (&big.Int{}).ModInverse(tmp, tk.N)
}

func (tk *ThresholdPublicKey) delta() *big.Int {
	return Factorial(tk.TotalNumberOfDecryptionServers)
}

func (tk *ThresholdPublicKey) verifyPartialDecryptions(shares []*PartialDecryption) error {
	if len(shares) < tk.Threshold {
		return errors.New("Threshold not meet")
	}
	tmp := make(map[int]bool)
	for _, share := range shares {
		tmp[share.Id] = true
	}
	if len(tmp) != len(shares) {
		return errors.New("two shares has been created by the same server")
	}
	return nil
}

func (tk *ThresholdPublicKey) updateLambda(share1, share2 *PartialDecryption, lambda *big.Int) *big.Int {
	num := new(big.Int).Mul(lambda, big.NewInt(int64(-share2.Id)))
	denom := big.NewInt(int64(share1.Id - share2.Id))
	return new(big.Int).Div(num, denom)
}

func (tk *ThresholdPublicKey) computeLambda(share *PartialDecryption, shares []*PartialDecryption) *big.Int {
	lambda := tk.delta()
	for _, share2 := range shares {
		if share2.Id != share.Id {
			lambda = tk.updateLambda(share, share2, lambda)
		}
	}
	return lambda
}

func (tk *ThresholdPublicKey) updateCprime(cprime, lambda *big.Int, share *PartialDecryption) *big.Int {
	twoLambda := new(big.Int).Mul(TWO, lambda)
	ret := tk.exp(share.Decryption, twoLambda, tk.GetNSquare())
	ret = new(big.Int).Mul(cprime, ret)
	return new(big.Int).Mod(ret, tk.GetNSquare())
}

func (tk *ThresholdPublicKey) exp(a, b, c *big.Int) *big.Int {
	if b.Cmp(ZERO) == -1 { // b < 0 ?
		ret := new(big.Int).Exp(a, new(big.Int).Neg(b), c)
		return new(big.Int).ModInverse(ret, c)
	}
	return new(big.Int).Exp(a, b, c)
}

func (tk *ThresholdPublicKey) computeDecryption(cprime *big.Int) *big.Int {
	l := L(cprime, tk.N)
	return new(big.Int).Mod(new(big.Int).Mul(tk.combineSharesConstant(), l), tk.N)
}

func (tk *ThresholdPublicKey) CombinePartialDecryptions(shares []*PartialDecryption) (*big.Int, error) {
	if err := tk.verifyPartialDecryptions(shares); err != nil {
		return nil, err
	}

	cprime := ONE
	for _, share := range shares {
		lambda := tk.computeLambda(share, shares)
		cprime = tk.updateCprime(cprime, lambda, share)
	}

	return tk.computeDecryption(cprime), nil
}

func (tk *ThresholdPublicKey) CombinePartialDecryptionsZKP(shares []*PartialDecryptionZKP) (*big.Int, error) {
	ret := make([]*PartialDecryption, 0)
	for _, share := range shares {
		if share.Verify() {
			ret = append(ret, &share.PartialDecryption)
		}
	}
	return tk.CombinePartialDecryptions(ret)
}

func (tk *ThresholdPublicKey) VerifyDecryption(encryptedMessage, decryptedMessage *big.Int, shares []*PartialDecryptionZKP) error {
	for _, share := range shares {
		if share.C.Cmp(encryptedMessage) != 0 {
			return errors.New("The encrypted message is not the same than the one in the shares")
		}
	}
	res, err := tk.CombinePartialDecryptionsZKP(shares)
	if err != nil {
		return err
	}
	if res.Cmp(decryptedMessage) != 0 {
		return errors.New("The decrypted message is not the same than the one in the shares")
	}
	return nil
}

type ThresholdPrivateKey struct {
	ThresholdPublicKey
	Id    int
	Share *big.Int
}

func (tpk *ThresholdPrivateKey) Decrypt(c *big.Int) *PartialDecryption {
	ret := new(PartialDecryption)
	ret.Id = tpk.Id
	exp := new(big.Int).Mul(tpk.Share, new(big.Int).Mul(TWO, tpk.delta()))
	gmpExp := gmp.NewInt(0).SetBytes(exp.Bytes())
	gmpC := gmp.NewInt(0).SetBytes(c.Bytes())
	gmpN2 := gmp.NewInt(0).SetBytes(tpk.GetNSquare().Bytes())
	ret.Decryption = big.NewInt(0).SetBytes(new(gmp.Int).Exp(gmpC, gmpExp, gmpN2).Bytes())
	return ret
}

func (tpk *ThresholdPrivateKey) copyVi() []*big.Int {
	ret := make([]*big.Int, len(tpk.Vi))
	for i, vi := range tpk.Vi {
		ret[i] = new(big.Int).Add(vi, big.NewInt(0))
	}
	return ret
}

func (tpk *ThresholdPrivateKey) getThresholdKey() *ThresholdPublicKey {
	ret := new(ThresholdPublicKey)
	ret.Threshold = tpk.Threshold
	ret.TotalNumberOfDecryptionServers = tpk.TotalNumberOfDecryptionServers
	ret.V = new(big.Int).Add(tpk.V, big.NewInt(0))
	ret.Vi = tpk.copyVi()
	ret.N = new(big.Int).Add(tpk.N, big.NewInt(0))
	return ret
}

func (tpk *ThresholdPrivateKey) computeZ(r, e *big.Int) *big.Int {
	tmp := new(big.Int).Mul(e, tpk.delta())
	tmp = new(big.Int).Mul(tmp, tpk.Share)
	return new(big.Int).Add(r, tmp)
}

func (tpk *ThresholdPrivateKey) computeHash(a, b, c4, ci2 *big.Int) *big.Int {
	hash := sha256.New()
	hash.Write(a.Bytes())
	hash.Write(b.Bytes())
	hash.Write(c4.Bytes())
	hash.Write(ci2.Bytes())
	return new(big.Int).SetBytes(hash.Sum([]byte{}))
}

func (tpk *ThresholdPrivateKey) DecryptAndProduceZKP(c *big.Int) (*PartialDecryptionZKP, error) {
	pd := new(PartialDecryptionZKP)
	pd.Key = tpk.getThresholdKey()
	pd.C = c
	pd.Id = tpk.Id
	pd.Decryption = tpk.Decrypt(c).Decryption

	r, err := rand.Int(rand.Reader, tpk.GetNSquare())
	if err != nil {
		return nil, err
	}
	c4 := new(big.Int).Exp(c, FOUR, nil)
	a := new(big.Int).Exp(c4, r, tpk.GetNSquare())

	b := new(big.Int).Exp(tpk.V, r, tpk.GetNSquare())

	ci2 := new(big.Int).Exp(pd.Decryption, big.NewInt(2), nil)

	pd.E = tpk.computeHash(a, b, c4, ci2)

	pd.Z = tpk.computeZ(r, pd.E)

	return pd, nil
}

func (tpk *ThresholdPrivateKey) Validate() error {
	m, err := rand.Int(rand.Reader, tpk.N)
	if err != nil {
		return err
	}
	c := tpk.Encrypt(m)
	if err != nil {
		return err
	}
	proof, err := tpk.DecryptAndProduceZKP(c.C)
	if err != nil {
		return err
	}
	if !proof.Verify() {
		return errors.New("invalid share.")
	}
	return nil
}

type PartialDecryption struct {
	Id         int
	Decryption *big.Int
}

type PartialDecryptionZKP struct {
	PartialDecryption
	Key *ThresholdPublicKey // the public key used to encrypt
	E   *big.Int            // the challenge
	Z   *big.Int            // the value needed to check to verify the decryption
	C   *big.Int            // the input cypher text
}

func (pd *PartialDecryptionZKP) verifyPart1() *big.Int {
	c4 := new(big.Int).Exp(pd.C, FOUR, nil)                  // c^4
	decryption2 := new(big.Int).Exp(pd.Decryption, TWO, nil) // c_i^2

	a1 := new(big.Int).Exp(c4, pd.Z, pd.Key.GetNSquare())          // (c^4)^Z
	a2 := new(big.Int).Exp(decryption2, pd.E, pd.Key.GetNSquare()) // (c_i^2)^E
	a2 = new(big.Int).ModInverse(a2, pd.Key.GetNSquare())
	a := new(big.Int).Mod(new(big.Int).Mul(a1, a2), pd.Key.GetNSquare())
	return a
}

func (pd *PartialDecryptionZKP) verifyPart2() *big.Int {
	vi := pd.Key.Vi[pd.Id-1]                                    // servers are indexed from 1
	b1 := new(big.Int).Exp(pd.Key.V, pd.Z, pd.Key.GetNSquare()) // V^Z
	b2 := new(big.Int).Exp(vi, pd.E, pd.Key.GetNSquare())       // (v_i)^E
	b2 = new(big.Int).ModInverse(b2, pd.Key.GetNSquare())
	b := new(big.Int).Mod(new(big.Int).Mul(b1, b2), pd.Key.GetNSquare())
	return b
}

func (pd *PartialDecryptionZKP) Verify() bool {
	a := pd.verifyPart1()
	b := pd.verifyPart2()
	hash := sha256.New()
	hash.Write(a.Bytes())
	hash.Write(b.Bytes())
	c4 := new(big.Int).Exp(pd.C, FOUR, nil)
	hash.Write(c4.Bytes())
	ci2 := new(big.Int).Exp(pd.Decryption, TWO, nil)
	hash.Write(ci2.Bytes())

	expectedE := new(big.Int).SetBytes(hash.Sum([]byte{}))
	return pd.E.Cmp(expectedE) == 0
}
