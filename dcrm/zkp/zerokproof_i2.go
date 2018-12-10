package zkp

import (
	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type Zkpi2 struct {
	U1 *Point
	U2 *big.Int
	U3 *big.Int
	Z1 *big.Int
	Z2 *big.Int
	S1 *big.Int
	S2 *big.Int
	T1 *big.Int
	T2 *big.Int
	T3 *big.Int
	E  *big.Int
	V1 *big.Int
	V3 *big.Int
}

//对签名数据进行核对
func (zkp *Zkpi2) Initialization(params *PublicParameters,
	eta1, eta2 *big.Int,
	//c *ECPoint,
	cx, cy,
	w, u, randomness *big.Int,
) {
	var N = params.PaillierPubKey.N
	var q = secp256k1.S256().N
	var nSquared = new(big.Int).Mul(N, N)
	var nTilde = params.NTilde
	var h1 = params.H1
	var h2 = params.H2
	var g = new(big.Int).Add(N, big.NewInt(1))

	//α ∈ (Z)q3
	var q2 = new(big.Int).Mul(q, q)
	var q3 = new(big.Int).Mul(q2, q)
	var alpha = RandomFromZn(q3)
	//β ∈ (Z)N
	var beta = RandomFromZn(N)
	//γ ∈ (Z)q3 * Ñ
	var gamma = RandomFromZn(new(big.Int).Mul(q3, nTilde))
	//δ ∈ (Z)q3
	//var delta= crypto.RandomFromZn(crypto.Pow(q, 3))
	//μ ∈ (Z) N
	var mu = RandomFromZnStar(N)
	//ν ∈ (Z)q3 * Ñ
	//var q3=crypto.Pow(q, 3)
	//var nu=crypto.RandomFromZn(new(big.Int).Mul(q3,NTilde))
	//θ ∈ (Z)q8
	var q6 = new(big.Int).Mul(q3, q3)
	var q8 = new(big.Int).Mul(q6, q2)
	var theta = RandomFromZn(q8)
	//τ ∈ (Z)q8 * Ñ
	var tau = RandomFromZn(new(big.Int).Mul(q8, nTilde))
	//ρ1 ∈ (Z)qÑ
	var rho1 = RandomFromZn(new(big.Int).Mul(q, nTilde))
	//ρ2 ∈ (Z)q6 * Ñ
	var rho2 = RandomFromZn(new(big.Int).Mul(q6, nTilde))

	//z1=(H1)eta1 * (H2)ρ1 mod Ñ
	hen := ModPowInsecure(h1, eta1, nTilde)
	hrn := ModPowInsecure(h2, rho1, nTilde)
	hh := new(big.Int).Mul(hen, hrn)
	zkp.Z1 = new(big.Int).Mod(hh, nTilde)

	//z2=(H1)eta2 * (H2)ρ2 mod Ñ
	hhen := ModPowInsecure(h1, eta2, nTilde)
	hhrn := ModPowInsecure(h2, rho2, nTilde)
	hhhh := new(big.Int).Mul(hhen, hhrn)
	zkp.Z2 = new(big.Int).Mod(hhhh, nTilde)

	//u1 = (g)α in G
	zkp.U1 = PointMul(alpha, &Point{cx, cy})

	//u2 = (Γ)α * (β)N mod N 2
	var gan = ModPowInsecure(g, alpha, nSquared)
	var bnn = ModPowInsecure(beta, N, nSquared)
	var gb = new(big.Int).Mul(gan, bnn)
	zkp.U2 = new(big.Int).Mod(gb, nSquared)

	//u3 = (H1)α * (H2)γ mod Ñ
	var han = ModPowInsecure(h1, alpha, nTilde)
	var hgn = ModPowInsecure(h2, gamma, nTilde)
	var hhag = new(big.Int).Mul(han, hgn)
	zkp.U3 = new(big.Int).Mod(hhag, nTilde)

	//v1 = (u)α (Γ)qθ * (μ)N mod N2
	var uan = ModPowInsecure(u, alpha, nSquared)
	var qtta = new(big.Int).Mul(q, theta)
	var gqn = ModPowInsecure(g, qtta, nSquared)
	var mnn = ModPowInsecure(mu, N, nSquared)
	var ug = new(big.Int).Mul(uan, gqn)
	var um = new(big.Int).Mul(ug, mnn)
	zkp.V1 = new(big.Int).Mod(um, nSquared)

	//v3 = (H1)θ * (H2)τ mod Ñ
	var hthn = ModPowInsecure(h1, theta, nTilde)
	var htnt = ModPowInsecure(h2, tau, nTilde)
	var aaa = new(big.Int).Mul(hthn, htnt)
	zkp.V3 = new(big.Int).Mod(aaa, nTilde)

	//e = hash(g, w, u, z1 , z2 , u1 , u2 , u3 , v1 , v3)
	digest := Sha256Hash(GetBytes(g), GetBytes(w), GetBytes(u), GetBytes(zkp.Z1),
		GetBytes(zkp.Z2), Get2Bytes(zkp.U1[0], zkp.U1[1]), GetBytes(zkp.U2), GetBytes(zkp.U3),
		GetBytes(zkp.V1), GetBytes(zkp.V3))
	if len(digest) == 0 {
		log.Crit("Assertion Error in zero knowledge proof i2")
	}
	zkp.E = new(big.Int).SetBytes(digest)

	//s1=e*eta1 + α
	eeta := new(big.Int).Mul(zkp.E, eta1)
	zkp.S1 = new(big.Int).Add(eeta, alpha)

	//s2=e*eta1 + γ
	erho := new(big.Int).Mul(zkp.E, rho1)
	zkp.S2 = new(big.Int).Add(erho, gamma)

	//t1=(rnd)e * μ mod N
	rande := ModPowInsecure(randomness, zkp.E, N)
	rmn := new(big.Int).Mul(rande, mu)
	zkp.T1 = new(big.Int).Mod(rmn, N)

	//t2=e*eta2 + θ
	eeta2 := new(big.Int).Mul(zkp.E, eta2)
	zkp.T2 = new(big.Int).Add(eeta2, theta)

	//t3=e*ρ2 + τ
	erho2 := new(big.Int).Mul(zkp.E, rho2)
	zkp.T3 = new(big.Int).Add(erho2, tau)
}

func (zkp *Zkpi2) Verify(params *PublicParameters, BitCurve *secp256k1.BitCurve,
	rx, ry,
	u *big.Int, w *big.Int) bool {

	h1 := params.H1
	h2 := params.H2
	N := params.PaillierPubKey.N
	nTilde := params.NTilde
	nSquared := new(big.Int).Mul(N, N)
	g := new(big.Int).Add(N, big.NewInt(1))
	q := secp256k1.S256().N
	var bitC = &ECPoint{
		X: secp256k1.S256().Gx,
		Y: secp256k1.S256().Gy,
	}
	finishedi2U1 := make(chan bool, 1)
	finishedi2U3 := make(chan bool, 1)
	finishedi2V1 := make(chan bool, 1)
	finishedi2V3 := make(chan bool, 1)
	finishedi2E := make(chan bool, 1)
	go zkp.checkU1(bitC.X, bitC.Y, rx, ry, finishedi2U1)
	go zkp.checkU3(h1, nTilde, h2, finishedi2U3)
	go zkp.checkV1(u, nSquared, q, g, N, w, finishedi2V1)
	go zkp.checkV3(h1, nTilde, h2, finishedi2V3)
	go zkp.checkE(bitC, w, u, g, finishedi2E)

	valueCheckPassed := 5
	for {
		select {
		case checkU1 := <-finishedi2U1:
			if checkU1 == false {
				log.Error("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(u1)")
				return false
			}
			log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(u1)")
			valueCheckPassed--
		case checkU3 := <-finishedi2U3:
			if checkU3 == false {
				log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(u3)")
				return false
			}
			log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(u3)")
			valueCheckPassed--
		case checkV1 := <-finishedi2V1:
			if checkV1 == false {
				log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(v1)")
				return false
			}
			log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(v1)")
			valueCheckPassed--
		case checkV3 := <-finishedi2V3:
			if checkV3 == false {
				log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(v3)")
				return false
			}
			log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(v3)")
			valueCheckPassed--
		case checkE := <-finishedi2E:
			if checkE == false {
				log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) failed when checking value(e)")
				return false
			}
			log.Info("[LOCK-OUT]Zero KnowLedge Proof(I2) Success when checking value(e)")
			valueCheckPassed--
		}

		if valueCheckPassed == 0 {
			break
		}
	}
	return true
}

//checkU1 check:u1=(c)s1 * (r)-e ing G
func (zkp *Zkpi2) checkU1(cx, cy *big.Int, rx, ry *big.Int, notify chan bool) {
	c := &Point{cx, cy}
	r := &Point{rx, ry}
	minuse := new(big.Int).Mul(zkp.E, big.NewInt(-1))
	minuse = new(big.Int).Mod(minuse, secp256k1.S256().N)
	u1 := pointAdd(PointMul(zkp.S1, c), PointMul(minuse, r))

	if u1[0].Cmp(zkp.U1[0]) == 0 && u1[1].Cmp(zkp.U1[1]) == 0 {
		notify <- true
		return
	} else {
		notify <- false
		return
	}
}

//checkU3 check: u3=(H1)s1 * (H2)s2 * (z1)-e mod Ñ
func (zkp *Zkpi2) checkU3(h1, nTilde, h2 *big.Int, notify chan bool) {
	hsn := ModPowInsecure(h1, zkp.S1, nTilde)
	hsnt := ModPowInsecure(h2, zkp.S2, nTilde)
	en := new(big.Int).Neg(zkp.E)
	hhss := new(big.Int).Mul(hsn, hsnt)
	zenn := ModPowInsecure(zkp.Z1, en, nTilde)
	hz := new(big.Int).Mul(hhss, zenn)
	hn := new(big.Int).Mod(hz, nTilde)
	cm := zkp.U3.Cmp(hn)

	if cm == 0 {
		notify <- true
	} else {
		notify <- false
	}
}

//checkV1 check:v1=(u)s1 * (τ)qt2 *(t1)N *(w)-e mod N2
func (zkp *Zkpi2) checkV1(u *big.Int, nSquared *big.Int, q *big.Int, g *big.Int, N *big.Int, w *big.Int, notify chan bool) {
	usn := ModPowInsecure(u, zkp.S1, nSquared)
	qt := new(big.Int).Mul(q, zkp.T2)
	gq := ModPowInsecure(g, qt, nSquared)
	tnn := ModPowInsecure(zkp.T1, N, nSquared)
	en := new(big.Int).Neg(zkp.E)
	wen := ModPowInsecure(w, en, nSquared)
	ug := new(big.Int).Mul(usn, gq)
	ugtnn := new(big.Int).Mul(ug, tnn)
	uw := new(big.Int).Mul(ugtnn, wen)
	uwn := new(big.Int).Mod(uw, nSquared)
	cm := zkp.V1.Cmp(uwn)

	if cm == 0 {
		notify <- true
		return
	} else {
		notify <- false
		return
	}
}

//checkV3 check:v3=(H1)t2 * (H2)t3 * (z2)-e mod Ñ
func (zkp *Zkpi2) checkV3(h1 *big.Int, nTilde *big.Int, h2 *big.Int, notify chan bool) {
	h1tn := ModPowInsecure(h1, zkp.T2, nTilde)
	htn := ModPowInsecure(h2, zkp.T3, nTilde)
	en := new(big.Int).Neg(zkp.E)
	zen := ModPowInsecure(zkp.Z2, en, nTilde)
	hh := new(big.Int).Mul(h1tn, htn)
	hz := new(big.Int).Mul(hh, zen)
	hzn := new(big.Int).Mod(hz, nTilde)
	cm := zkp.V3.Cmp(hzn)

	if cm == 0 {
		notify <- true
		return
	} else {
		notify <- false
		return
	}
}

//checkE check:e = hash(g, w, u, z1 , z2 , u1 , u2 , u3 , v1 , v3 )  no v2
func (zkp *Zkpi2) checkE(bitC *ECPoint, w, u, g *big.Int, notify chan bool) { //Get2Bytes(bitC.X,bitC.Y)
	var result = Sha256Hash(GetBytes(g), GetBytes(w), GetBytes(u),
		GetBytes(zkp.Z1), GetBytes(zkp.Z2), Get2Bytes(zkp.U1[0], zkp.U1[1]),
		GetBytes(zkp.U2), GetBytes(zkp.U3),
		GetBytes(zkp.V1), GetBytes(zkp.V3))

	if zkp.E.Cmp(new(big.Int).SetBytes(result)) == 0 {
		notify <- true
		return
	} else {
		notify <- false
		return
	}
}
