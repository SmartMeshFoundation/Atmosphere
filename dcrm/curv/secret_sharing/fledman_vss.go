package secret_sharing

import (
	"math/big"

	"fmt"

	"crypto/ecdsa"

	"encoding/hex"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var S = secp256k1.S256()
var BigOne = big.NewInt(1)

//公钥
type GE struct {
	X *big.Int
	Y *big.Int
}

func NewGE(x, y *big.Int) *GE {
	return &GE{
		X: new(big.Int).Set(x),
		Y: new(big.Int).Set(y),
	}
}
func (g *GE) Clone() *GE {
	return &GE{
		X: new(big.Int).Set(g.X),
		Y: new(big.Int).Set(g.Y),
	}
}

type ShamirSecretSharing struct {
	Threshold  int
	ShareCount int
}

type VerifiableSS struct {
	Parameters  ShamirSecretSharing
	Commitments []*GE
}

func Share(t, n int, secret *big.Int) (*VerifiableSS, []*big.Int) {
	poly := SamplePolynomial(t, secret)
	var index []int
	for i := 1; i <= n; i++ {
		index = append(index, i)
	}
	secretShares := EvaluatePolynomial(poly, index)
	log.Trace(fmt.Sprintf("secretShares=%s", secretShares))
	var commitments []*GE
	for _, p := range poly {
		x, y := S.ScalarBaseMult(p.Bytes())
		commitments = append(commitments, &GE{x, y})
	}

	return &VerifiableSS{
		Parameters: ShamirSecretSharing{
			Threshold:  t,
			ShareCount: n,
		},
		Commitments: commitments,
	}, secretShares
}
func SamplePolynomial(t int, coef0 *big.Int) []*big.Int {
	var bs []*big.Int
	bs = append(bs, coef0)
	for i := 0; i < t; i++ {
		k, _ := crypto.GenerateKey()
		bs = append(bs, k.D)
	}
	bs = []*big.Int{
		coef0,
		big.NewInt(1),
		big.NewInt(2),
		big.NewInt(3),
	}
	return bs
}

//s1=s1+s2 mod N
func ModAdd(s1, s2 *big.Int) *big.Int {
	s1.Mod(s1, S.N)
	s2 = new(big.Int).Mod(s2, S.N)
	s1.Add(s1, s2)
	s1.Mod(s1, S.N)
	return s1
}

//s1=s1*s2 mod N
func ModMul(s1, s2 *big.Int) *big.Int {
	s1.Mod(s1, S.N)
	s2 = new(big.Int).Mod(s2, S.N)
	s1.Mul(s1, s2)
	s1.Mod(s1, S.N)
	return s1
}

//s1=s1-s2 mod N
func ModSub(s1, s2 *big.Int) *big.Int {
	return modSubInternal(s1, s2, S.N)
}
func modSubInternal(s1, s2, modulus *big.Int) *big.Int {
	s1.Mod(s1, modulus)
	s2 = new(big.Int).Mod(s2, modulus)
	if s1.Cmp(s2) >= 0 {
		s1.Sub(s1, s2)
		return s1.Mod(s1, modulus)
	}
	big0 := big.NewInt(0)
	s2 = big0.Sub(big0, s2)
	s2.Add(s2, modulus)
	s1.Add(s1, s2)
	s1.Mod(s1, modulus)
	return s1
}

//s1=s1**s2 mod N
func ModPow(s1, s2 *big.Int) *big.Int {
	return s1.Exp(s1, s2, S.N)
}
func EvaluatePolynomial(coefficients []*big.Int, index []int) []*big.Int {
	var bs []*big.Int
	for i := 0; i < len(index); i++ {
		point := big.NewInt(int64(index[i]))
		point = point.Mod(point, S.N)
		log.Trace(fmt.Sprintf("point=%s", point.Text(16)))
		log.Trace(fmt.Sprintf("coefficients=%s", utils.StringInterface(coefficients, 3)))
		sum := new(big.Int).Set(coefficients[len(coefficients)-1])
		for j := len(coefficients) - 2; j >= 0; j-- {
			log.Trace(fmt.Sprintf("sum=%s,coef=%s", sum.Text(16), coefficients[j].Text(16)))
			ModMul(sum, point)
			ModAdd(sum, coefficients[j])
		}
		bs = append(bs, sum)
	}
	return bs
}
func PointAdd(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	key1 := &ecdsa.PublicKey{
		Curve: S,
		X:     x1,
		Y:     y1,
	}
	key2 := &ecdsa.PublicKey{
		Curve: S,
		X:     x2,
		Y:     y2,
	}
	key1bin := crypto.FromECDSAPub(key1)
	key2bin := crypto.FromECDSAPub(key2)
	k, err := secp256k1.AddPoint(key1bin, key2bin)
	if err != nil {
		panic(fmt.Sprintf("add err %s", err))
	}
	k2, err := crypto.UnmarshalPubkey(k)
	if err != nil {
		panic(fmt.Sprintf("UnmarshalPubkey err %s", err))
	}
	return k2.X, k2.Y
}
func PointSub(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	order := new(big.Int).Set(S.P)
	minusY := new(big.Int).Set(order)
	x = x2
	y = y2
	minusY = modSubInternal(minusY, y, order)
	return PointAdd(x1, y1, x2, minusY)

}
func Strtoxy(s string) (x, y *big.Int) {
	s1 := s[:64]
	s2 := s[64:]
	s1b, _ := hex.DecodeString(s1)
	s2b, _ := hex.DecodeString(s2)
	s1bc := make([]byte, len(s1b))
	s2bc := make([]byte, len(s1b))
	i := 0
	for j := len(s1b) - 1; j >= 0; j-- {
		s1bc[i] = s1b[j]
		s2bc[i] = s2b[j]
		i++
	}
	x = new(big.Int)
	x.SetBytes(s1bc)
	y = new(big.Int)
	y.SetBytes(s2bc)
	return
}
func Xytostr(x, y *big.Int) string {
	x1 := utils.BigIntTo32Bytes(x)
	y1 := utils.BigIntTo32Bytes(y)
	x2 := make([]byte, len(x1))
	y2 := make([]byte, len(x1))
	i := 0
	for j := len(x1) - 1; j >= 0; j-- {
		x2[i] = x1[j]
		y2[i] = y1[j]
		i++
	}
	s := fmt.Sprintf("%s%s", hex.EncodeToString(x2), hex.EncodeToString(y2))
	return s
}
func (v *VerifiableSS) ValidateShare(secretShare *big.Int, index int) bool {
	x, y := S.ScalarBaseMult(secretShare.Bytes())
	ssPoint := &GE{x, y}
	indexFe := big.NewInt(int64(index))
	indexFe = indexFe.Mod(indexFe, S.N)
	l := len(v.Commitments)
	log.Trace(fmt.Sprintf("indexfe=%s", indexFe))
	head := v.Commitments[l-1]
	for j := l - 2; j >= 0; j-- {
		c := v.Commitments[j]
		log.Trace(fmt.Sprintf("acc=%s,x=%s", Xytostr(head.X, head.Y), Xytostr(c.X, c.Y)))
		x, y = S.ScalarMult(head.X, head.Y, indexFe.Bytes())
		log.Trace(fmt.Sprintf("t=%s", Xytostr(x, y)))
		x, y = PointAdd(x, y, c.X, c.Y)
		//log.Trace(fmt.Sprintf("x1=%s,y1=%s", x.Text(16), y.Text(16)))
		//x, y = S.Add(head.X, head.Y, x, y)
		log.Trace(fmt.Sprintf("after add %s", Xytostr(x, y)))
		head = &GE{x, y}
	}
	log.Trace(fmt.Sprintf("sspoint=%s,commit_to_point=%s", utils.StringInterface(ssPoint, 3), utils.StringInterface(head, 3)))
	return ssPoint.X.Cmp(head.X) == 0 && ssPoint.Y.Cmp(head.Y) == 0
}

func (v *VerifiableSS) MapShareToNewParams(index int, s []int) *big.Int {
	if len(s) <= v.reconstructLimit() {
		panic("reconstructLimit")
	}
	var points []*big.Int
	for i := 1; i <= v.Parameters.ShareCount; i++ {
		p := big.NewInt(int64(i))
		p = p.Mod(p, S.N)
		points = append(points, p)
	}
	xi := points[index]
	num := big.NewInt(1) //不需要mod了吧?
	denum := big.NewInt(1)
	for i := 0; i < len(s); i++ {
		if s[i] != index {
			num = ModMul(num, points[s[i]])
		}
	}
	for i := 0; i < len(s); i++ {
		if s[i] != index {
			xj := new(big.Int).Set(points[s[i]])
			ModSub(xj, xi)
			ModMul(denum, xj)
		}
	}
	log.Trace(fmt.Sprintf("num=%s,denum=%s", num.Text(16), denum.Text(16)))
	denum = Invert(denum, S.N)
	ModMul(num, denum)
	return num
}

/*
   // Performs a Lagrange interpolation in field Zp at the origin
   // for a polynomial defined by `points` and `values`.
   // `points` and `values` are expected to be two arrays of the same size, containing
   // respectively the evaluation points (x) and the value of the polynomial at those point (p(x)).

   // The result is the value of the polynomial at x=0. It is also its zero-degree coefficient.

   // This is obviously less general than `newton_interpolation_general` as we
   // only get a single value, but it is much faster.
*/
func lagrange_interpolation_at_zero(points []*big.Int, values []*big.Int) *big.Int {
	var lagCoef []*big.Int
	for i := 0; i < len(values); i++ {
		xi := points[i]
		yi := values[i]
		num := BigInt2PrivateKey(big.NewInt(1))
		denum := BigInt2PrivateKey(big.NewInt(1))
		for j := 0; j < len(values); j++ {
			if i != j {
				num = ModMul(num, points[j])
			}
		}
		for j := 0; j < len(values); j++ {
			if i != j {
				//denum*=points[j]-xi
				xj := new(big.Int).Set(points[j])
				ModSub(xj, xi)
				ModMul(denum, xj)
			}
		}
		denum = Invert(denum, S.N)
		ModMul(num, denum)
		ModMul(num, yi) //num*denum*yi
		lagCoef = append(lagCoef, num)
	}
	var result = new(big.Int).Set(lagCoef[0])
	for i := 1; i < len(values); i++ {
		ModAdd(result, lagCoef[i])
	}
	return result
}
func (v *VerifiableSS) reconstructLimit() int {
	return v.Parameters.Threshold + 1
}

func (v *VerifiableSS) Reconstruct(indices []int, shares []*big.Int) *big.Int {
	if len(shares) != len(indices) {
		panic("arg error")
	}
	if len(shares) != v.reconstructLimit() {
		panic("arg error")
	}
	var points []*big.Int
	for _, i := range indices {
		b := big.NewInt(int64(i + 1))
		b = BigInt2PrivateKey(b)
		points = append(points, b)
	}
	return lagrange_interpolation_at_zero(points, shares)
}

// Invert calculates the inverse of k in GF(P) using Fermat's method.
// This has better constant-time properties than Euclid's method (implemented
// in math/big.Int.ModInverse) although math/big itself isn't strictly
// constant-time so it's not perfect.
func Invert(k, N *big.Int) *big.Int {
	two := big.NewInt(2)
	nMinus2 := new(big.Int).Sub(N, two)
	return new(big.Int).Exp(k, nMinus2, N)
}

func str2bigint(s string) *big.Int {
	i := new(big.Int)
	i.SetString(s, 16)
	return i
}

func RandomPrivateKey() *big.Int {
	key, _ := crypto.GenerateKey()
	return key.D
}

func BigInt2PrivateKey(i *big.Int) *big.Int {
	b := new(big.Int).Set(i)
	b.Mod(b, S.N)
	return b
}
