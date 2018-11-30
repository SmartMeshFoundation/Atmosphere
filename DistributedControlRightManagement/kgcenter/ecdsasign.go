package kgcenter

import (
	"math/big"

	"bytes"

	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/sirupsen/logrus"
)

type ECDSASignature struct {
	r               *big.Int
	s               *big.Int
	recoveryParam   int32
	roudFiveAborted bool
}

func (ecds *ECDSASignature) New() {
}

func (ecds *ECDSASignature) New2(r *big.Int, s *big.Int) {
	ecds.r = r
	ecds.s = s
}

func (ecds *ECDSASignature) New3(r *big.Int, s *big.Int, recoveryParam int32) {
	ecds.r = r
	ecds.s = s
	ecds.recoveryParam = recoveryParam
}

func (ecds *ECDSASignature) ToBytes() []byte {
	buf := new(bytes.Buffer)
	buf.Write(utils.BigIntTo32Bytes(ecds.r))
	buf.Write(utils.BigIntTo32Bytes(ecds.s))
	buf.Write([]byte{byte(ecds.recoveryParam)})
	return buf.Bytes()
}

func (ecds *ECDSASignature) verify(message string, pkx *big.Int, pky *big.Int) bool {

	z, _ := new(big.Int).SetString(message, 16)
	ss := new(big.Int).ModInverse(ecds.s, secp256k1.S256().N)
	zz := new(big.Int).Mul(z, ss) //v*r
	u1 := new(big.Int).Mod(zz, secp256k1.S256().N)

	zz2 := new(big.Int).Mul(ecds.r, ss)
	u2 := new(big.Int).Mod(zz2, secp256k1.S256().N)

	if u1.Sign() == -1 {
		u1.Add(u1, secp256k1.S256().P)
	}
	ug := make([]byte, 32)
	math.ReadBits(u1, ug[:])
	ugx, ugy := KMulG(ug[:])

	if u2.Sign() == -1 {
		u2.Add(u2, secp256k1.S256().P)
	}
	upk := make([]byte, 32)
	math.ReadBits(u2, upk[:])
	upkx, upky := secp256k1.S256().ScalarMult(pkx, pky, upk[:])

	xxx, _ := secp256k1.S256().Add(ugx, ugy, upkx, upky)
	xR := new(big.Int).Mod(xxx, secp256k1.S256().N)

	if xR.Cmp(ecds.r) == 0 {
		logrus.Info("ECDSA Signature Verify Passed! (r,s) is a Valid Siganture!")
		logrus.Info("(r,s) r=", ecds.r)
		logrus.Info("(r,s) s=", ecds.s)
		return true
	}

	logrus.Error("ERROR: ECDSA Signature Verify NOT Passed! ")
	return false
}

func KMulG(k []byte) (*big.Int, *big.Int) {
	return secp256k1.S256().ScalarBaseMult(k)
}

func (ecds *ECDSASignature) getRoudFiveAborted() bool {
	return ecds.roudFiveAborted
}

func (ecds *ECDSASignature) setRoudFiveAborted(roudFiveAborted bool) {
	ecds.roudFiveAborted = roudFiveAborted
}

func (ecds *ECDSASignature) GetR() *big.Int {
	return ecds.r
}

func (ecds *ECDSASignature) setR(r *big.Int) {
	ecds.r = r
}

func (ecds *ECDSASignature) GetS() *big.Int {
	return ecds.s
}

func (ecds *ECDSASignature) setS(s *big.Int) {
	ecds.s = s
}

func (ecds *ECDSASignature) GetRecoveryParam() int32 {
	return ecds.recoveryParam
}

func (ecds *ECDSASignature) setRecoveryParam(recoveryParam int32) {
	ecds.recoveryParam = recoveryParam
}
