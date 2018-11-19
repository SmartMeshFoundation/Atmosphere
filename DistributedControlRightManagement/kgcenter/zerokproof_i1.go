package kgcenter

import (
	"math/big"
	"math/rand"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/sirupsen/logrus"
)

type Zkpi1 struct {
	z  *big.Int
	u1 *big.Int
	u2 *big.Int
	s1 *big.Int
	s2 *big.Int
	s3 *big.Int
	e  *big.Int
	v  *big.Int
}

var (
	finishedi1U1=make(chan bool,1)
	finishedi1U2=make(chan bool,1)
	finishedi1V=make(chan bool,1)
	finishedi1E=make(chan bool,1)
)

//对证名人身份的核对
func (zkp *Zkpi1)Initialization(params *PublicParameters,
	eta *big.Int,
	rand *rand.Rand,
	r ,c1 ,c2 ,c3 *big.Int,
) {
	//(x)y均表示幂运算：x为底数 y为指数
	var N= params.paillierPubKey.N
	var q= secp256k1.S256().N
	var nSquared= new(big.Int).Mul(N, N)
	var nTilde= params.nTilde
	var h1= params.h1
	var h2= params.h2
	var g= new(big.Int).Add(N, big.NewInt(1))

	var q2=new(big.Int).Mul(q,q)
	var q3=new(big.Int).Mul(q2,q)
	var alpha= RandomFromZn(q3)
	var beta= RandomFromZn(N)
	var gamma= RandomFromZn(new(big.Int).Mul(q3, nTilde))
	var rho= RandomFromZn(new(big.Int).Mul(q, nTilde))

	//证明人计算(注：普利斯顿大学的资料上z和u1做反了):
	//z = (h1)η * (h2)ρ  mod Ñ
	var mx1= ModPowInsecure(h1, eta, nTilde)
	var mx2= ModPowInsecure(h2, rho, nTilde)
	var mx12= new(big.Int).Mul(mx1, mx2)
	zkp.z = new(big.Int).Mod(mx12, nTilde)
	//u1 = (Γ)α * (β)N mod N2
	var my1= ModPowInsecure(g, alpha, nSquared)
	var my2= ModPowInsecure(beta, N, nSquared)
	var my12= new(big.Int).Mul(my1, my2)
	zkp.u1 = new(big.Int).Mod(my12, nSquared)
	//u2 = (h1)α * (h2)γ mod Ñ
	var mz1= ModPowInsecure(h1, alpha, nTilde)
	var mz2= ModPowInsecure(h2, gamma, nTilde)
	var mz12= new(big.Int).Mul(mz1, mz2)
	zkp.u2 = new(big.Int).Mod(mz12, nTilde)
	//v = (c2)α mod N2
	zkp.v = ModPowInsecure(c2, alpha, nSquared)

	digest := Sha256Hash(GetBytes(c1), GetBytes(c2), GetBytes(c3),
		GetBytes(zkp.z), GetBytes(zkp.u1), GetBytes(zkp.u2), GetBytes(zkp.v))
	if len(digest) == 0 {
		logrus.Panic("Assertion Error in zero knowledge proof i1")
	}
	//e = hash(c1 , c2 , c3 , z, u1 , u2 , v)
	zkp.e = new(big.Int).SetBytes(digest)
	//s1 = eη + α
	var ee = new(big.Int).Mul(zkp.e, eta)
	zkp.s1 = new(big.Int).Add(ee, alpha)
	//s2 = (r)e×β mod N
	var ren = ModPowInsecure(r, zkp.e, N)
	var renb = new(big.Int).Mul(ren, beta)
	zkp.s2 = new(big.Int).Mod(renb, N)
	//s3 = eρ + γ
	var er = new(big.Int).Mul(zkp.e, rho)
	zkp.s3 = new(big.Int).Add(er, gamma)
}

func (zkp *Zkpi1)verify(params *PublicParameters,
	CURVE *secp256k1.BitCurve,
	c1,c2,c3 *big.Int,
) bool {
	var h1= params.h1
	var h2= params.h2
	var N= params.paillierPubKey.N
	var nTilde= params.nTilde
	var nSquared= new(big.Int).Mul(N,N)
	var g= new(big.Int).Add(N, big.NewInt(1))

	valueCheckPassed := 4

	go zkp.checkU1(g, nSquared, N, c3)
	go zkp.checkU2(h1, nTilde, h2)
	go zkp.checkV(c2, nSquared, c1)
	go zkp.checkE(c1, c2, c3)

	for {
		select {
		case checkU1 := <-finishedi1U1:
			if checkU1 == false {
				logrus.Error("Zero KnowLedge Proof(I1) failed when checking value(u1)")
				return false
			}
			logrus.Info("Zero KnowLedge Proof(I1) Success when checking value(u1)")
			valueCheckPassed--
		case checkU2 := <-finishedi1U2:
			if checkU2 == false {
				logrus.Error("Zero KnowLedge Proof(I1) failed when checking value(u2)")
				return false
			}
			logrus.Info("Zero KnowLedge Proof(I1) Success when checking value(u2)")
			valueCheckPassed--
		case checkV := <-finishedi1V:
			if checkV == false {
				logrus.Error("Zero KnowLedge Proof(I1) failed when checking value(v)")
				return false
			}
			logrus.Info("Zero KnowLedge Proof(I1) Success when checking value(v)")
			valueCheckPassed--
		case checkE := <-finishedi1E:
			if checkE == false {
				logrus.Error("Zero KnowLedge Proof(I1) failed when checking value(e)")
				return false
			}
			logrus.Error("Zero KnowLedge Proof(I1) Success when checking value(e)")
			valueCheckPassed--
		}
		if valueCheckPassed == 0 {
			break
		}
	}
	return true

}

//z = (Γ)s1 * (s2)N * (c3)-e mod N2
func (zkp *Zkpi1)checkU1(g ,nSquared,N ,c3 *big.Int) {
	var x= ModPowInsecure(g, zkp.s1, nSquared)
	var y = ModPowInsecure(zkp.s2, N, nSquared)
	var c3neg= new(big.Int).Neg(zkp.e)
	var z = ModPowInsecure(c3, c3neg, nSquared)
	var mulxy = new(big.Int).Mul(x, y)
	var mulxyz = new(big.Int).Mul(mulxy, z)
	var result = new(big.Int).Mod(mulxyz, nSquared)
	if zkp.u1.Cmp(result) == 0 {
		finishedi1U1 <- true
	} else {
		finishedi1U1 <- false
	}
}

//u2 = (h1)s1 * (h2)s3 * (u1)−e mod Ñ
func (zkp *Zkpi1)checkU2(h1 ,nTilde ,h2 *big.Int) {
	var x = ModPowInsecure(h1, zkp.s1, nTilde)
	var y = ModPowInsecure(h2, zkp.s3, nTilde)
	var eneg = new(big.Int).Neg(zkp.e)
	var z = ModPowInsecure(zkp.z, eneg, nTilde)
	var mulxy = new(big.Int).Mul(x, y)
	var mulxyz = new(big.Int).Mul(mulxy, z)
	var result = new(big.Int).Mod(mulxyz, nTilde)
	if zkp.u2.Cmp(result) == 0 {
		finishedi1U2 <- true

	} else {
		finishedi1U2 <- false
	}
}

//v = (c2)s1 * (c1)−e mod N2
func (zkp *Zkpi1)checkV(c2 ,nSquared ,c1 *big.Int)  {
	var x = ModPowInsecure(c2,zkp.s1,nSquared)
	var eneg = new(big.Int).Neg(zkp.e)
	var y = ModPowInsecure(c1,eneg,nSquared)
	var mulxy = new(big.Int).Mul(x,y)
	var result = new(big.Int).Mod(mulxy,nSquared)
	if zkp.v.Cmp(result) == 0 {
		finishedi1V <-true
	}else {
		finishedi1V <-false
	}
}

//e = hash(c1 , c2 , c3 , z, u1 , u2 ,v)
func (zkp *Zkpi1)checkE(c1 ,c2 ,c3 *big.Int) {
	var result = Sha256Hash(GetBytes(c1), GetBytes(c2), GetBytes(c3),
		GetBytes(zkp.z), GetBytes(zkp.u1), GetBytes(zkp.u2), GetBytes(zkp.v), )
	if zkp.e.Cmp(new(big.Int).SetBytes(result)) == 0 {
		finishedi1E <- true
	} else {
		finishedi1E <- false
	}
}