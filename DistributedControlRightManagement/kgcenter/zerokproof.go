package kgcenter

import (
	"math/big"
	"math/rand"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/common/math"
)

type Zkp struct {
	z *big.Int
	u1 *big.Int
	u2 *big.Int
	u3 *big.Int
	e *big.Int
	s1 *big.Int
	s2 *big.Int
	s3 *big.Int
}
/*
The prover sends all of these values to the Verifier. The Verifier checks that
all the values are in the correct range and moreover that the following equations
hold:
u1 = (g)s1 * (y)−e in G
u2 = (Γ)s1 * (s2)N * (w)−e mod (N)2
u3 = (h1)s1 *(h2)s3 * (z)−e mod Ñ
e = hash(g,y,w,z,u1,u2,u3)
*/
var (
	finishedU1=make(chan bool,1)
	finishedU2=make(chan bool,1)
	finishedU3=make(chan bool,1)
	finishedE=make(chan bool,1)
)

//g的定义需要查找原文献资料，是否为1，对曲线的影响是什么
func (zkp *Zkp) Initialization(params *PublicParameters,
	eta *big.Int,
	rand *rand.Rand,
	cx,cy,w ,r *big.Int,
)  {
	//(x)y均表示幂运算：x为底数 y为指数
	var N= params.paillierPubKey.N
	var q= secp256k1.S256().N
	var nSquared= new(big.Int).Mul(N, N)
	var nTilde= params.nTilde
	var h1= params.h1
	var h2= params.h2

	var g= new(big.Int).Add(N, big.NewInt(1))

	//α ∈ (Z)q3
	var q2=new(big.Int).Mul(q,q)
	var q3=new(big.Int).Mul(q2,q)
	var alpha= RandomFromZn(q3)
	//β ∈ (Z)N
	var beta= RandomFromZn(N)
	//ρ 2 ∈ (Z)q * Ñ
	var rho=RandomFromZn(new(big.Int).Mul(q,nTilde))
	//γ ∈ (Z)q3 * Ñ
	var gamma= RandomFromZn(new(big.Int).Mul(q3, nTilde))

	//证明人计算:
	//z = (h1)η * (h2)ρ * mod Ñ
	var mx1=ModPowInsecure(h1,eta,nTilde)
	var mx2=ModPowInsecure(h2,rho,nTilde)
	var mx12=new(big.Int).Mul(mx1,mx2)
	zkp.z=new(big.Int).Mod(mx12,nTilde)
	if alpha.Sign() == -1 {
		alpha.Add(alpha,secp256k1.S256().P)
	}
	//u1 = (g)α in G (IsOnCurve)  //todo
	var alpha256=make([]byte,256/8)
	math.ReadBits(alpha,alpha256[:])
	zkp.u1=new(big.Int).Mul(new(big.Int).SetBytes(Get2Bytes(cx,cy)),alpha)
	var u1256=make([]byte,256/8)
	math.ReadBits(zkp.u1,u1256)
	zkp.u1=new(big.Int).Rsh(zkp.u1,1024)
	//u2 = (Γ)α * (β)N mod (N)2
	var my1=ModPowInsecure(g,alpha,nSquared)
	var my2=ModPowInsecure(beta,N,nSquared)
	var my12=new(big.Int).Mul(my1,my2)
	zkp.u2=new(big.Int).Mod(my12,nSquared)
	//u3 = (h1)α * (h2)γ mod N
	var mz1=ModPowInsecure(h1,alpha,nTilde)
	var mz2=ModPowInsecure(h2,gamma,nTilde)
	var mz12=new(big.Int).Mul(mz1,mz2)
	zkp.u3=new(big.Int).Mod(mz12,nTilde)
	//e = hash(g, y, w, z, u1 , u2 , u3)
/*	digest := Sha256Hash(GetBytes(g), Get2Bytes(y.X,y.Y), GetBytes((w)),
		GetBytes((zkp.z)), Get2Bytes(zkp.u1.X,zkp.u1.Y), GetBytes(zkp.u2), GetBytes(zkp.u3))*/
	digest := Sha256Hash(GetBytes(g),Get2Bytes(cx,cy), GetBytes((w)),
		GetBytes((zkp.z)), GetBytes(zkp.u1), GetBytes(zkp.u2), GetBytes(zkp.u3))//Get2Bytes(zkp.u1x,zkp.u1y)
	if len(digest) == 0 {
		logrus.Panic("Assertion Error in zero knowledge proof when lock-in progress")
	}
	zkp.e = new(big.Int).SetBytes(digest)

	//s1 = eη + α
	var ee = new(big.Int).Mul(zkp.e,eta)
	zkp.s1 = new(big.Int).Add(ee,alpha)
	//s2 = (r)e *β mod N
	var re =ModPowInsecure(r,zkp.e,N)//k
	var rb = new(big.Int).Mul(re,beta)
	zkp.s2 = new(big.Int).Mod(rb,N)
	//s3 = eρ + γ
	var er = new(big.Int).Mul(zkp.e,rho)
	zkp.s3 = new(big.Int).Add(er,gamma)
}

func (zkp *Zkp)Verify(params *PublicParameters,rx,ry,
	w *big.Int,
) bool {
	var h1= params.h1
	var h2= params.h2
	var N= params.paillierPubKey.N
	var nTilde= params.nTilde
	var nSquared= new(big.Int).Mul(N,N)
	var g= new(big.Int).Add(N, big.NewInt(1))
	var bitC=&ECPoint{
		X:secp256k1.S256().Gx,
		Y:secp256k1.S256().Gy,
	}
	valueCheckPassed := 4

	go zkp.checkU1(bitC.X,bitC.Y,rx,ry,nTilde)
	go zkp.checkU2(g ,nSquared,N ,w)
	go zkp.checkU3(h1 ,nTilde ,h2)
	go zkp.checkE(bitC ,w,g)

	for {
		select {
		case checkU1 := <-finishedU1:
			if checkU1 == false {
				logrus.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(u1)")
				return false
			}
			logrus.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(u1)")
			valueCheckPassed--
		case checkU2 := <-finishedU2:
			if checkU2 == false {
				logrus.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(u2)")
				return false
			}
			logrus.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(u2)")
			valueCheckPassed--
		case checkV := <-finishedU3:
			if checkV == false {
				logrus.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(u3)")
				return false
			}
			logrus.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(u3)")
			valueCheckPassed--
		case checkE := <-finishedE:
			if checkE == false {
				logrus.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(e)")
				return false
			}
			logrus.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(e)")
			valueCheckPassed--
		}
		if valueCheckPassed == 0 {
			break
		}
	}
	return true
}

//check u1 = (g)s1 * (y)−e in G(IsOnCurve default true)  =>|g*s1 + (y)*−e -u1|=0,1
func (zkp *Zkp)checkU1(bx,by, rx,ry,nTilde *big.Int) {
	x1 := new(big.Int).Mul(new(big.Int).SetBytes(Get2Bytes(bx, by)), zkp.s1)
	//fmt.Println(configs.G.P.BitLen())
	var nege = new(big.Int).Neg(zkp.e)
	x2 := new(big.Int).Mul(new(big.Int).SetBytes(Get2Bytes(rx, ry)), nege)
	result := new(big.Int).Add(x1, x2)
	var result256=make([]byte,256/8)
	math.ReadBits(result,result256)
	result=new(big.Int).Rsh(result,1024)
	//fmt.Println("result",result)
	//fmt.Println("zkp.u1",zkp.u1)
	subReuslt:=new(big.Int).Sub(result,zkp.u1)
	//fmt.Println("bb:",subReuslt)
	subReuslt=new(big.Int).Abs(subReuslt)
	if subReuslt.Cmp(big.NewInt(0))==0 ||subReuslt.Cmp(big.NewInt(1))==0{
		finishedU1 <- true
		return
	} else {
		finishedU1 <- false
		return
	}
}

//check u2 = (Γ)s1 * (s2)N * (w)−e mod (N)2
func (zkp *Zkp)checkU2(g ,nSquared,N ,w *big.Int) {
	var x = ModPowInsecure(g, zkp.s1, nSquared)
	var y= ModPowInsecure(zkp.s2, N, nSquared)
	var mulxy= new(big.Int).Mul(x, y)
	var c3neg = new(big.Int).Neg(zkp.e)
	var z= ModPowInsecure(w, c3neg, nSquared)
	var mulxyz= new(big.Int).Mul(mulxy, z)
	var result= new(big.Int).Mod(mulxyz, nSquared)
	/*	logrus.Info("result=",result)
	logrus.Info("u2=",zkp.u2)*/
	if zkp.u2.Cmp(result) == 0 {
		finishedU2 <- true
		return
	} else {
		finishedU2 <- false
		return
	}
}

//check u3 = (h1)s1 *(h2)s3 * (z)−e mod Ñ
func (zkp *Zkp)checkU3(h1 ,nTilde ,h2 *big.Int) {
	var x = ModPowInsecure(h1, zkp.s1, nTilde)
	var y = ModPowInsecure(h2, zkp.s3, nTilde)
	var mulxy = new(big.Int).Mul(x, y)
	var eneg = new(big.Int).Neg(zkp.e)
	var z = ModPowInsecure(zkp.z, eneg, nTilde)
	var mulxyz = new(big.Int).Mul(mulxy, z)
	var result = new(big.Int).Mod(mulxyz, nTilde)
	if zkp.u3.Cmp(result) == 0 {
		finishedU3 <- true
		return

	} else {
		finishedU3 <- false
		return
	}
}

//check e = hash(g,y,w,z,u1,u2,u3)
func (zkp *Zkp)checkE(bitC *ECPoint ,w,g *big.Int) {
	var result = Sha256Hash(GetBytes(g),Get2Bytes(bitC.X,bitC.Y), GetBytes(w),
		GetBytes(zkp.z), GetBytes(zkp.u1), GetBytes(zkp.u2), GetBytes(zkp.u3))//Get2Bytes(zkp.u1x,zkp.u1y)
	if zkp.e.Cmp(new(big.Int).SetBytes(result)) == 0 {
		finishedE <- true
		return
	} else {
		finishedE <- false
		return
	}
}