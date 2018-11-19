package kgcenter

import (
	"math/big"
	"math/rand"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/common/math"
)

type Zkpi2 struct {
	u1 *big.Int
	u2 *big.Int
	u3 *big.Int
	z1 *big.Int
	z2 *big.Int
	s1 *big.Int
	s2 *big.Int
	t1 *big.Int
	t2 *big.Int
	t3 *big.Int
	e  *big.Int
	v1 *big.Int
	v3 *big.Int
}

var (
	finishedi2U1=make(chan bool,1)
	finishedi2U3=make(chan bool,1)
	finishedi2V1=make(chan bool,1)
	finishedi2V3=make(chan bool,1)
	finishedi2E=make(chan bool,1)
)

//对签名数据进行核对
func (zkp *Zkpi2)Initialization(params *PublicParameters,
	eta1,eta2 *big.Int,
	rand *rand.Rand,
	//c *ECPoint,
	cx,cy,
	w,u ,randomness *big.Int,
){
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
	//γ ∈ (Z)q3 * Ñ
	var gamma= RandomFromZn(new(big.Int).Mul(q3, nTilde))
	//δ ∈ (Z)q3
	//var delta= crypto.RandomFromZn(crypto.Pow(q, 3))
	//μ ∈ (Z) N
	var mu=RandomFromZnStar(N)
	//ν ∈ (Z)q3 * Ñ
	//var q3=crypto.Pow(q, 3)
	//var nu=crypto.RandomFromZn(new(big.Int).Mul(q3,nTilde))
	//θ ∈ (Z)q8
	var q6 = new(big.Int).Mul(q3,q3)
	var q8 = new(big.Int).Mul(q6,q2)
	var theta=RandomFromZn(q8)
	//τ ∈ (Z)q8 * Ñ
	var tau=RandomFromZn(new(big.Int).Mul(q8,nTilde))
	//ρ1 ∈ (Z)qÑ
	var rho1=RandomFromZn(new(big.Int).Mul(q,nTilde))
	//ρ 2 ∈ (Z)q6 * Ñ
	var rho2=RandomFromZn(new(big.Int).Mul(q6,nTilde))
	//z1
	hen := ModPowInsecure(h1,eta1,nTilde)
	hrn := ModPowInsecure(h2,rho1,nTilde)
	hh := new(big.Int).Mul(hen,hrn)
	zkp.z1 = new(big.Int).Mod(hh,nTilde)
	//z2
	hhen := ModPowInsecure(h1,eta2,nTilde)
	hhrn := ModPowInsecure(h2,rho2,nTilde)
	hhhh := new(big.Int).Mul(hhen,hhrn)
	zkp.z2 = new(big.Int).Mod(hhhh,nTilde)
	//证明人计算:
	alpha256 := make([]byte, 32)
	math.ReadBits(alpha,alpha256[:])
	zkp.u1=new(big.Int).Mul(new(big.Int).SetBytes(Get2Bytes(cx,cy)),alpha)
	var u1256=make([]byte,256/8)
	math.ReadBits(zkp.u1,u1256)
	zkp.u1=new(big.Int).Rsh(zkp.u1,1024)
	//u2 = (Γ)α * (β)N mod N 2
	var gan = ModPowInsecure(g,alpha,nSquared)
	var bnn = ModPowInsecure(beta,N,nSquared)
	var gb = new(big.Int).Mul(gan,bnn)
	zkp.u2 = new(big.Int).Mod(gb,nSquared)
	//u3 = (h1)α * (h2)γ mod Ñ
	var han = ModPowInsecure(h1,alpha,nTilde)
	var hgn = ModPowInsecure(h2,gamma,nTilde)
	var hhag = new(big.Int).Mul(han,hgn)
	zkp.u3 = new(big.Int).Mod(hhag,nTilde)
	//v1 = (u)α (Γ)qθ * (μ)N mod N2
	var uan = ModPowInsecure(u,alpha,nSquared)
	var qtta = new(big.Int).Mul(q,theta)
	var gqn = ModPowInsecure(g,qtta,nSquared)
	var mnn = ModPowInsecure(mu,N,nSquared)
	var ug = new(big.Int).Mul(uan,gqn)
	var um = new(big.Int).Mul(ug,mnn)
	zkp.v1 = new(big.Int).Mod(um,nSquared)
	//v3 = (h1)θ * (h2)τ mod Ñ
	var hthn = ModPowInsecure(h1,theta, nTilde)
	var htnt = ModPowInsecure(h2,tau,nTilde)
	var aaa = new(big.Int).Mul(hthn,htnt)
	zkp.v3 = new(big.Int).Mod(aaa,nTilde)
	//e = hash(g, w, u, z1 , z2 , u1 , u2 , u3 , v1 , v2 , v3)
	digest := Sha256Hash(GetBytes(g),GetBytes(w),GetBytes(u),GetBytes(zkp.z1),
		GetBytes(zkp.z2),GetBytes(zkp.u1),GetBytes(zkp.u2),GetBytes(zkp.u3),
		GetBytes(zkp.v1),GetBytes(zkp.v3))//Get2Bytes(cx,cy),
	if len(digest)==0{
		logrus.Panic("Assertion Error in zero knowledge proof i2")
	}
	zkp.e=new(big.Int).SetBytes(digest)
	eeta := new(big.Int).Mul(zkp.e,eta1) //e.multiply(eta1)
	zkp.s1 = new(big.Int).Add(eeta,alpha)

	erho := new(big.Int).Mul(zkp.e,rho1)
	zkp.s2 = new(big.Int).Add(erho,gamma)

	rande := ModPowInsecure(randomness,zkp.e,N)
	rmn := new(big.Int).Mul(rande,mu)
	zkp.t1 = new(big.Int).Mod(rmn,N)

	eeta2 := new(big.Int).Mul(zkp.e,eta2)
	zkp.t2 = new(big.Int).Add(eeta2,theta)

	erho2 := new(big.Int).Mul(zkp.e,rho2)
	zkp.t3 = new(big.Int).Add(erho2,tau)
}

func (zkp *Zkpi2) verify(params *PublicParameters,BitCurve *secp256k1.BitCurve,
	//r *ECPoint,
	rx,ry,
	u *big.Int,w *big.Int) bool {

	h1 := params.h1
	h2 := params.h2
	N := params.paillierPubKey.N
	nTilde := params.nTilde
	nSquared := new(big.Int).Mul(N,N)
	g := new(big.Int).Add(N, big.NewInt(1))
	q := secp256k1.S256().N
	var bitC=&ECPoint{
		X:secp256k1.S256().Gx,
		Y:secp256k1.S256().Gy,
	}
	//r:=&ECPoint{rx,ry}
	go zkp.checkU1(bitC.X,bitC.Y,rx,ry)
	go zkp.checkU3(h1,nTilde,h2)
	go zkp.checkV1(u,nSquared,q,g,N,w)
	go zkp.checkV3(h1,nTilde,h2)
	go zkp.checkE(bitC,w,u,g)

	valueCheckPassed := 5
	for {
		select {
		case checkU1 := <-finishedi2U1:
			if checkU1 == false {
				logrus.Error("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(u1)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(u1)")
			valueCheckPassed--
		case checkU3 := <-finishedi2U3:
			if checkU3 == false {
				logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(u3)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(u3)")
			valueCheckPassed--
		case checkV1 := <-finishedi2V1:
			if checkV1 == false {
				logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(v1)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(v1)")
			valueCheckPassed--
		case checkV3 := <-finishedi2V3:
			if checkV3 == false {
				logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(v3)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(v3)")
			valueCheckPassed--
		case checkE := <-finishedi2E:
			if checkE == false {
				logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(e)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(e)")
			valueCheckPassed--
		}

		if valueCheckPassed == 0 {
			break
		}
	}
	return true
}

func (zkp *Zkpi2) checkU1(cx,cy *big.Int,rx,ry *big.Int) {

	x1 := new(big.Int).Mul(new(big.Int).SetBytes(Get2Bytes(cx, cx)), zkp.s1)
	var nege = new(big.Int).Neg(zkp.e)
	x2 := new(big.Int).Mul(new(big.Int).SetBytes(Get2Bytes(rx, rx)), nege)
	result := new(big.Int).Add(x1, x2)
	var result256=make([]byte,256/8)
	math.ReadBits(result,result256)
	result=new(big.Int).Rsh(result,1024)
	subReuslt:=new(big.Int).Sub(result,zkp.u1)
	subReuslt=new(big.Int).Abs(subReuslt)
	if subReuslt.Cmp(big.NewInt(0))==0 ||subReuslt.Cmp(big.NewInt(1))==0{
		finishedi2U1 <- true
		return
	} else {
		finishedi2U1 <- false
		return
	}
}

func (zkp *Zkpi2) checkU3(h1 ,nTilde,h2 *big.Int) {
	hsn := ModPowInsecure(h1, zkp.s1, nTilde)
	hsnt := ModPowInsecure(h2, zkp.s2, nTilde)
	en := new(big.Int).Neg(zkp.e)
	hhss := new(big.Int).Mul(hsn, hsnt)
	zenn := ModPowInsecure(zkp.z1, en, nTilde)

	hz := new(big.Int).Mul(hhss, zenn)
	hn := new(big.Int).Mod(hz, nTilde)
	cm := zkp.u3.Cmp(hn)
	if cm == 0 {
		finishedi2U3 <- true
	} else {
		finishedi2U3 <- false
	}
}

func (zkp *Zkpi2) checkV1(u *big.Int,nSquared *big.Int,q *big.Int,g *big.Int,N *big.Int,w *big.Int) {
	usn := ModPowInsecure(u, zkp.s1, nSquared)
	qt := new(big.Int).Mul(q, zkp.t2)
	gq := ModPowInsecure(g, qt, nSquared)
	tnn := ModPowInsecure(zkp.t1, N, nSquared)
	en := new(big.Int).Neg(zkp.e)
	wen := ModPowInsecure(w, en, nSquared)
	ug := new(big.Int).Mul(usn, gq)
	ugtnn := new(big.Int).Mul(ug, tnn)
	uw := new(big.Int).Mul(ugtnn, wen)
	uwn := new(big.Int).Mod(uw, nSquared)
	cm := zkp.v1.Cmp(uwn)
	if cm == 0 {
		finishedi2V1 <- true
	} else {
		finishedi2V1 <- false
	}
}

func (zkp *Zkpi2)checkV3(h1 *big.Int,nTilde *big.Int,h2 *big.Int) {
	h1tn := ModPowInsecure(h1, zkp.t2, nTilde)
	htn := ModPowInsecure(h2, zkp.t3, nTilde)
	en := new(big.Int).Neg(zkp.e)
	zen := ModPowInsecure(zkp.z2, en, nTilde)
	hh := new(big.Int).Mul(h1tn, htn)
	hz := new(big.Int).Mul(hh, zen)
	hzn := new(big.Int).Mod(hz, nTilde)
	cm := zkp.v3.Cmp(hzn)

	if cm == 0 {
		finishedi2V3 <- true
	} else {
		finishedi2V3 <- false
	}
}

//e = hash(g, w, u, z1 , z2 , u1 , u2 , u3 , v1 , v2 , v3 )  无v2
func (zkp *Zkpi2) checkE(bitC *ECPoint ,w,u,g *big.Int) {//Get2Bytes(bitC.X,bitC.Y)
	var result = Sha256Hash(GetBytes(g),GetBytes(w),GetBytes(u),
		GetBytes(zkp.z1), GetBytes(zkp.z2),
		GetBytes(zkp.u1), GetBytes(zkp.u2), GetBytes(zkp.u3),
		GetBytes(zkp.v1),GetBytes(zkp.v3),)
	if zkp.e.Cmp(new(big.Int).SetBytes(result)) == 0 {
		finishedi2E <- true
		return
	} else {
		finishedi2E <- false
		return
	}
}


