package kgcenter

import (
	"testing"
	"math/big"
)

func TestGcd(t *testing.T) {
 i:=Gcd(big.NewInt(255),big.NewInt(25))
 t.Log(i)
}