package kgcenter

import (
	"crypto/sha256"
	"math/big"
	"math/rand"
	"time"
	//"github.com/Roasbeef/go-go-gadget-paillier"
	crand "crypto/rand"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func ModPowInsecure(base, exponent, modulus *big.Int) *big.Int {
	if exponent.Cmp(big.NewInt(0)) >= 0 {
		return new(big.Int).Exp(base, exponent, modulus)
	}
	derivatives := new(big.Int).ModInverse(base, modulus)
	exp := new(big.Int).Neg(exponent)
	return new(big.Int).Exp(derivatives, exp, modulus)
}

func Pow(x *big.Int, n int) *big.Int {
	if n == 0 {
		return big.NewInt(1)
	} else {
		return x.Mul(x, Pow(x, n-1))
	}

}

func GetBytes(n *big.Int) []byte {
	nlen := (n.BitLen() + 7) / 8
	newBuffer := make([]byte, nlen)
	math.ReadBits(n, newBuffer)
	return newBuffer
}

func Get2Bytes(n1, n2 *big.Int) []byte {
	n1Len := (n1.BitLen() + 7) / 8
	n2Len := (n1.BitLen() + 7) / 8
	newLen := n1Len + n2Len
	newBuffer := make([]byte, newLen)
	math.ReadBits(n1, newBuffer[0:n1Len])
	math.ReadBits(n2, newBuffer[n2Len:])
	return newBuffer
}

func Sha256Hash(inputs ...[]byte) []byte {
	messageDigest := sha256.New()
	for i := range inputs {
		messageDigest.Write([]byte(inputs[i]))
	}
	return messageDigest.Sum([]byte{})
}

func isProbablePrime(num *big.Int) bool {
	return num.ProbablyPrime(0)
}

/*func GenerateParams(k,kPrime int,rand *rand.Rand,paillierPubKey  *paillier.PublicKey) (pp *PublicParameters) {
	//var primeCertainty = k
	var p,q,pPrime,qPrime,pPrimeqPrime,nHat *big.Int
	for {
		//BigInteger(int bitLength, int certainty, Random rnd) BigInteger(kPrime / 2, primeCertainty, rand);
		p,_=crand.Prime(crand.Reader,kPrime / 2)
		ps:=new(big.Int).Sub(p,big.NewInt(1))
		pPrime=new(big.Int).Div(ps,big.NewInt(2))
		if isProbablePrime(pPrime)==true{
			break
		}
	}
	for{
		q,_=crand.Prime(crand.Reader,kPrime / 2)
		qs:=new(big.Int).Sub(q,big.NewInt(1))
		qPrime=new(big.Int).Div(qs,big.NewInt(2))
		if isProbablePrime(qPrime)==true{
			break
		}
		break
	}
	// generate nhat. the product of two safe primes, each of length
	//	kPrime/2

	nHat=new(big.Int).Mul(p,q)
	h2:=RandomFromZnStar(nHat)
	pPrimeqPrime=new(big.Int).Mul(pPrime,qPrime)
	x:=RandomFromZn(pPrimeqPrime)
	h1:=ModPowInsecure(h2,x,nHat)
	CURVE:=secp256k1.S256()

	pp=new(PublicParameters)
	pp.Initialization(CURVE,nHat,kPrime,h1,h2,paillierPubKey)
	return
}*/

func GenerateParams(BitCurve *secp256k1.BitCurve, primeCertainty int32, kPrime int32, rnd *rand.Rand, paillierPubKey *PublicKey) *PublicParameters {
	var p, q, pPrime, qPrime, pPrimeqPrime, nHat *big.Int
	for {
		p, _ = crand.Prime(crand.Reader, int(kPrime/2))
		psub := new(big.Int).Sub(p, big.NewInt(1))
		pPrime = new(big.Int).Div(psub, big.NewInt(2))
		if isProbablePrime(pPrime) == true {
			break
		}
	}

	for {
		q, _ = crand.Prime(crand.Reader, int(kPrime/2))
		qsub := new(big.Int).Sub(q, big.NewInt(1))
		qPrime = new(big.Int).Div(qsub, big.NewInt(2))
		if isProbablePrime(qPrime) == true {
			break
		}
	}

	nHat = new(big.Int).Mul(p, q)
	h2 := RandomFromZnStar(nHat)
	pPrimeqPrime = new(big.Int).Mul(pPrime, qPrime)
	x := RandomFromZn(pPrimeqPrime)
	h1 := ModPowInsecure(h2, x, nHat)
	pparms := new(PublicParameters)
	pparms.Initialization(BitCurve, nHat, kPrime, h1, h2, paillierPubKey)
	return pparms
}

//随机性地返回一个数（ Z_n^*）
func RandomFromZnStar(n *big.Int) *big.Int {
	var result *big.Int
	for {
		xRnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		traget := big.NewInt(1)
		traget.Lsh(traget, uint(n.BitLen())) //左移n.BitLen位
		result = new(big.Int).Rand(xRnd, traget)
		if result.Cmp(n) != -1 {
			break
		}
	}
	return result
}

//生成一个随机数，范围在(UnixNano,256位的最大整数（256个1）)，取一个比G的阶N小的一个随机数
func RandomFromZn(p *big.Int) *big.Int {
	var result *big.Int
	for {
		xRnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		traget := big.NewInt(1)
		traget.Lsh(traget, uint(p.BitLen())) //左移n.BitLen位
		result = new(big.Int).Rand(xRnd, traget)
		if result.Cmp(p) < 0 {
			break
		}
	}
	return result
}

//求最大公约数
func Gcd(x, y *big.Int) *big.Int {
	var tmp *big.Int
	for {
		tmp = new(big.Int).Mod(x, y)
		if tmp.Cmp(big.NewInt(1)) != -1 {
			x = y
			y = tmp
		} else {
			return y
		}
	}
}

type Point [2]*big.Int

func PointMul(scalar *big.Int, p *Point) *Point {
	r := &Point{}
	for i := 0; i < scalar.BitLen(); i++ {
		if scalar.Bit(i) == 1 {
			r = pointAdd(r, p)
		}
		p = pointAdd(p, p)
	}
	return r
}

var p = new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFC, 0x2F})

func pointAdd(p1, p2 *Point) *Point {
	if Isinfinite(p1) {
		return p2
	}
	if Isinfinite(p2) {
		return p1
	}
	if (p1[0].Cmp(p2[0]) == 0) && (p1[1].Cmp(p2[1]) != 0) {
		return &Point{}
	}
	var lmbda *big.Int
	if (p1[0].Cmp(p2[0]) == 0) && (p1[1].Cmp(p2[1]) == 0) {
		// 3 * x1 * x1 * (2 * y1)^(p - 2) mod p
		lmbda = mod(mul(big.NewInt(3), p1[0], p1[0],
			exp(mul(big.NewInt(2), p1[1]), sub(p, big.NewInt(2)))),
			p)
	} else {
		// (y2 - y1) * (x2 - x1)^(p - 2) mod p
		lmbda = mod(mul(sub(p2[1], p1[1]),
			exp(sub(p2[0], p1[0]), sub(p, big.NewInt(2)))),
			p)
	}
	x3 := mod(sub(sub(mul(lmbda, lmbda), p1[0]), p2[0]), p)

	return &Point{x3,
		mod(sub(mul(lmbda, sub(p1[0], x3)), p1[1]), p),
	}
}

func mod(x, y *big.Int) *big.Int {
	return new(big.Int).Mod(x, y)
}
func add(x, y *big.Int) *big.Int {
	return new(big.Int).Add(x, y)
}
func exp(x, y *big.Int) *big.Int {
	return new(big.Int).Exp(x, y, p)
}
func sub(x, y *big.Int) *big.Int {
	return new(big.Int).Sub(x, y)
}
func mul(x ...*big.Int) *big.Int {
	m := big.NewInt(1)
	for _, xi := range x {
		m = new(big.Int).Mul(m, xi)
	}
	return m
}

//p点是否在无穷远
func Isinfinite(P *Point) bool {
	if P[1] == nil || P[1].Cmp(big.NewInt(0)) == 0 {
		return true
	}
	return false
}
