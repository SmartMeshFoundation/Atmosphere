package kgcenter

import (
	"math/big"
	"testing"
)

func TestGcd(t *testing.T) {
	i := Gcd(big.NewInt(255), big.NewInt(25))
	t.Log(i)
}
