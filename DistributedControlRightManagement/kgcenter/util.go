package kgcenter

import (
	"math/big"
	"math/rand"
	"time"
	"crypto/sha256"
	"github.com/fusion/common/math"
	//"github.com/Roasbeef/go-go-gadget-paillier"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	crand "crypto/rand"
)

func ModPowInsecure(base,exponent,modulus *big.Int) *big.Int{
	if exponent.Cmp(big.NewInt(0))>=0{
		return new(big.Int).Exp(base,exponent,modulus)
	}
	derivatives:=new(big.Int).ModInverse(base,modulus)
	exp:=new(big.Int).Neg(exponent)
	return new(big.Int).Exp(derivatives,exp,modulus)
}

func Pow(x *big.Int,n int) *big.Int  {
	if n==0{
		return big.NewInt(1)
	}else {
		return x.Mul(x,Pow(x,n-1))
	}

}


func GetBytes(n *big.Int) []byte {
	nlen:=(n.BitLen()+7)/8
	newBuffer:=make([]byte,nlen)
	math.ReadBits(n,newBuffer)
	return newBuffer
}

func Get2Bytes(n1,n2 *big.Int) []byte {
	n1Len:=(n1.BitLen()+7)/8
	n2Len:=(n1.BitLen()+7)/8
	newLen:=n1Len+n2Len
	newBuffer:=make([]byte,newLen)
	math.ReadBits(n1,newBuffer[0:n1Len])
	math.ReadBits(n2,newBuffer[n2Len:])
	return newBuffer
}

func Sha256Hash(inputs ...[]byte ) []byte {
	messageDigest := sha256.New()
	for i := range inputs {
		messageDigest.Write([]byte(inputs[i]))
	}
	return messageDigest.Sum([]byte{})
}

func isProbablePrime(num *big.Int) bool {
	return  num.ProbablyPrime(0)
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

func GenerateParams(BitCurve *secp256k1.BitCurve,primeCertainty int32,kPrime int32,rnd *rand.Rand,paillierPubKey *PublicKey) *PublicParameters {
	var p,q,pPrime,qPrime,pPrimeqPrime,nHat *big.Int
	for {
		p,_ = crand.Prime(crand.Reader,int(kPrime / 2))

		one,_ := new(big.Int).SetString("1",10)
		psub := new(big.Int).Sub(p,one)
		two,_ := new(big.Int).SetString("2",10)
		pPrime = new(big.Int).Div(psub,two)
		if isProbablePrime(pPrime) == true {
			break
		}
	}

	for {
		q,_ = crand.Prime(crand.Reader,int(kPrime / 2))

		one,_ := new(big.Int).SetString("1",10)
		qsub := new(big.Int).Sub(q,one)
		two,_ := new(big.Int).SetString("2",10)
		qPrime = new(big.Int).Div(qsub,two)
		if isProbablePrime(qPrime) == true {
			break
		}
	}

	nHat = new(big.Int).Mul(p,q)
	h2 := RandomFromZnStar(nHat)
	pPrimeqPrime = new(big.Int).Mul(pPrime,qPrime)
	x := RandomFromZn(pPrimeqPrime)
	h1 := ModPowInsecure(h2,x,nHat)
	pparms := new(PublicParameters)
	pparms.Initialization(BitCurve,nHat,kPrime, h1, h2, paillierPubKey)
	return pparms
}

//随机性地返回一个数（ Z_n^*）
func RandomFromZnStar(n *big.Int) *big.Int{
	var result *big.Int
	for {
		xRnd:=rand.New(rand.NewSource(time.Now().UnixNano()))
		traget:=big.NewInt(1)
		traget.Lsh(traget,uint(n.BitLen()))//左移n.BitLen位
		result=new(big.Int).Rand(xRnd,traget)
		if result.Cmp(n)!=-1 {
			break
		}
	}
	return result
}

//生成一个随机数，范围在(UnixNano,256位的最大整数（256个1）)，取一个比G的阶N小的一个随机数
func RandomFromZn(p *big.Int ) *big.Int{
	var result *big.Int
	for {
		xRnd:=rand.New(rand.NewSource(time.Now().UnixNano()))
		traget:=big.NewInt(1)
		traget.Lsh(traget,uint(p.BitLen()))//左移n.BitLen位
		result=new(big.Int).Rand(xRnd,traget)
		if result.Cmp(p)<0{
			break
		}
	}
	return result
}

//求最大公约数
func Gcd(x, y *big.Int) *big.Int {
	var tmp *big.Int
	for {
		tmp = new(big.Int).Mod(x,y)
		if tmp.Cmp(big.NewInt(1))!=-1 {
			x = y
			y = tmp
		} else {
			return y
		}
	}
}

