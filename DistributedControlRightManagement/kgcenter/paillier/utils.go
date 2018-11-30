package paillier

import (
	"crypto/rand"
	"io"
	"math/big"
)

var ZERO = big.NewInt(0)
var ONE = big.NewInt(1)
var TWO = big.NewInt(2)
var FOUR = big.NewInt(4)

func Factorial(n int) *big.Int {
	ret := big.NewInt(1)
	for i := 1; i <= n; i++ {
		ret = new(big.Int).Mul(ret, big.NewInt(int64(i)))
	}
	return ret
}

func GetRandomNumberInMultiplicativeGroup(n *big.Int, random io.Reader) (*big.Int, error) {
	r, err := rand.Int(random, n)
	if err != nil {
		return nil, err
	}
	zero := big.NewInt(0)
	one := big.NewInt(1)
	if zero.Cmp(r) == 0 || one.Cmp(new(big.Int).GCD(nil, nil, n, r)) != 0 {
		return GetRandomNumberInMultiplicativeGroup(n, random)
	}
	return r, nil

}

func GetRandomGeneratorOfTheQuadraticResidue(n *big.Int, rand io.Reader) (*big.Int, error) {
	r, err := GetRandomNumberInMultiplicativeGroup(n, rand)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Mod(new(big.Int).Mul(r, r), n), nil
}
