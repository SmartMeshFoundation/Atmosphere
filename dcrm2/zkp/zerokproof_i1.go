package zkp

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/sirupsen/logrus"
)

type Zkpi1 struct {
	Z  *big.Int
	U1 *big.Int
	U2 *big.Int
	S1 *big.Int
	S2 *big.Int
	S3 *big.Int
	E  *big.Int
	V  *big.Int
}

func (zkp *Zkpi1) Initialization(params *PublicParameters,
	eta *big.Int,
	r, c1, c2, c3 *big.Int,
) {
	var N = params.PaillierPubKey.N
	var q = secp256k1.S256().N
	var nSquared = new(big.Int).Mul(N, N)
	var nTilde = params.NTilde
	var h1 = params.H1
	var h2 = params.H2
	var g = new(big.Int).Add(N, big.NewInt(1))

	var q2 = new(big.Int).Mul(q, q)
	var q3 = new(big.Int).Mul(q2, q)
	var alpha = RandomFromZn(q3)
	var beta = RandomFromZn(N)
	var gamma = RandomFromZn(new(big.Int).Mul(q3, nTilde))
	var rho = RandomFromZn(new(big.Int).Mul(q, nTilde))

	//被证明人计算(注：普利斯顿大学的资料上z和u1做反了):
	//z = (H1)η * (H2)ρ  mod Ñ
	var mx1 = ModPowInsecure(h1, eta, nTilde)
	var mx2 = ModPowInsecure(h2, rho, nTilde)
	var mx12 = new(big.Int).Mul(mx1, mx2)
	zkp.Z = new(big.Int).Mod(mx12, nTilde)

	//u1 = (Γ)α * (β)N mod N2
	var my1 = ModPowInsecure(g, alpha, nSquared)
	var my2 = ModPowInsecure(beta, N, nSquared)
	var my12 = new(big.Int).Mul(my1, my2)
	zkp.U1 = new(big.Int).Mod(my12, nSquared)

	//u2 = (H1)α * (H2)γ mod Ñ
	var mz1 = ModPowInsecure(h1, alpha, nTilde)
	var mz2 = ModPowInsecure(h2, gamma, nTilde)
	var mz12 = new(big.Int).Mul(mz1, mz2)
	zkp.U2 = new(big.Int).Mod(mz12, nTilde)

	//v = (c2)α mod N2
	zkp.V = ModPowInsecure(c2, alpha, nSquared)

	//e = hash(c1 , c2 , c3 , z, u1 , u2 , v)
	digest := Sha256Hash(GetBytes(c1), GetBytes(c2), GetBytes(c3),
		GetBytes(zkp.Z), GetBytes(zkp.U1), GetBytes(zkp.U2), GetBytes(zkp.V))
	if len(digest) == 0 {
		logrus.Panic("Assertion Error in zero knowledge proof i1")
	}
	zkp.E = new(big.Int).SetBytes(digest)

	//s1 = eη + α
	var ee = new(big.Int).Mul(zkp.E, eta)
	zkp.S1 = new(big.Int).Add(ee, alpha)

	//s2 = (r)e×β mod N
	var ren = ModPowInsecure(r, zkp.E, N)
	var renb = new(big.Int).Mul(ren, beta)
	zkp.S2 = new(big.Int).Mod(renb, N)

	//S3 = eρ + γ
	var er = new(big.Int).Mul(zkp.E, rho)
	zkp.S3 = new(big.Int).Add(er, gamma)
}

func (zkp *Zkpi1) Verify(params *PublicParameters, CURVE *secp256k1.BitCurve, c1, c2, c3 *big.Int) bool {
	var h1 = params.H1
	var h2 = params.H2
	var N = params.PaillierPubKey.N
	var nTilde = params.NTilde
	var nSquared = new(big.Int).Mul(N, N)
	var g = new(big.Int).Add(N, big.NewInt(1))

	valueCheckPassed := 4
	finishedi1U1 := make(chan bool, 1)
	finishedi1U2 := make(chan bool, 1)
	finishedi1V := make(chan bool, 1)
	finishedi1E := make(chan bool, 1)

	go zkp.checkU1(g, nSquared, N, c3, finishedi1U1)
	go zkp.checkU2(h1, nTilde, h2, finishedi1U2)
	go zkp.checkV(c2, nSquared, c1, finishedi1V)
	go zkp.checkE(c1, c2, c3, finishedi1E)

	for {
		select {
		case checkU1 := <-finishedi1U1:
			if checkU1 == false {
				logrus.Error("[LOCK-OUT]Zero KnowLedge Proof(I1) failed when checking value(u1)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I1) Success when checking value(u1)")
			valueCheckPassed--
		case checkU2 := <-finishedi1U2:
			if checkU2 == false {
				logrus.Error("[LOCK-OUT]Zero KnowLedge Proof(I1) failed when checking value(u2)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I1) Success when checking value(u2)")
			valueCheckPassed--
		case checkV := <-finishedi1V:
			if checkV == false {
				logrus.Error("[LOCK-OUT]Zero KnowLedge Proof(I1) failed when checking value(v)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I1) Success when checking value(v)")
			valueCheckPassed--
		case checkE := <-finishedi1E:
			if checkE == false {
				logrus.Error("[LOCK-OUT]Zero KnowLedge Proof(I1) failed when checking value(e)")
				return false
			}
			logrus.Info("[LOCK-OUT]Zero KnowLedge Proof(I1) Success when checking value(e)")
			valueCheckPassed--
		}
		if valueCheckPassed == 0 {
			break
		}
	}
	return true

}

//checkU1 check u1= (Γ)s1 * (s2)N * (c3)-e mod N2
func (zkp *Zkpi1) checkU1(g, nSquared, N, c3 *big.Int, notify chan bool) {
	var x = ModPowInsecure(g, zkp.S1, nSquared)
	var y = ModPowInsecure(zkp.S2, N, nSquared)
	var c3neg = new(big.Int).Neg(zkp.E)
	var z = ModPowInsecure(c3, c3neg, nSquared)
	var mulxy = new(big.Int).Mul(x, y)
	var mulxyz = new(big.Int).Mul(mulxy, z)
	var result = new(big.Int).Mod(mulxyz, nSquared)
	if zkp.U1.Cmp(result) == 0 {
		notify <- true
	} else {
		notify <- false
	}
}

//checkU2 check:u2 = (H1)s1 * (H2)S3 * (u1)−e mod Ñ
func (zkp *Zkpi1) checkU2(h1, nTilde, h2 *big.Int, notify chan bool) {
	var x = ModPowInsecure(h1, zkp.S1, nTilde)
	var y = ModPowInsecure(h2, zkp.S3, nTilde)
	var eneg = new(big.Int).Neg(zkp.E)
	var z = ModPowInsecure(zkp.Z, eneg, nTilde)
	var mulxy = new(big.Int).Mul(x, y)
	var mulxyz = new(big.Int).Mul(mulxy, z)
	var result = new(big.Int).Mod(mulxyz, nTilde)
	if zkp.U2.Cmp(result) == 0 {
		notify <- true

	} else {
		notify <- false
	}
}

//checkV check:v = (c2)s1 * (c1)−e mod N2
func (zkp *Zkpi1) checkV(c2, nSquared, c1 *big.Int, notify chan bool) {
	var x = ModPowInsecure(c2, zkp.S1, nSquared)
	var eneg = new(big.Int).Neg(zkp.E)
	var y = ModPowInsecure(c1, eneg, nSquared)
	var mulxy = new(big.Int).Mul(x, y)
	var result = new(big.Int).Mod(mulxy, nSquared)
	if zkp.V.Cmp(result) == 0 {
		notify <- true
	} else {
		notify <- false
	}
}

//checkE check:e = hash(c1 , c2 , c3 , z, u1 , u2 ,v)
func (zkp *Zkpi1) checkE(c1, c2, c3 *big.Int, notify chan bool) {
	var result = Sha256Hash(GetBytes(c1), GetBytes(c2), GetBytes(c3),
		GetBytes(zkp.Z), GetBytes(zkp.U1), GetBytes(zkp.U2), GetBytes(zkp.V))
	if zkp.E.Cmp(new(big.Int).SetBytes(result)) == 0 {
		notify <- true
	} else {
		notify <- false
	}
}
