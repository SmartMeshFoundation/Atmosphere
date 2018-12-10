package zkp

import (
	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type Zkp struct {
	Z  *big.Int
	U1 *Point
	U2 *big.Int
	U3 *big.Int
	E  *big.Int
	S1 *big.Int
	S2 *big.Int
	S3 *big.Int
}

func (zkp *Zkp) Initialization(params *PublicParameters,
	eta *big.Int,
	cx, cy, w, r *big.Int,
) {
	//(x)y均表示幂运算：x为底数 y为指数
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
	//ρ 2 ∈ (Z)q * Ñ
	var rho = RandomFromZn(new(big.Int).Mul(q, nTilde))
	//γ ∈ (Z)q3 * Ñ
	var gamma = RandomFromZn(new(big.Int).Mul(q3, nTilde))
	//被证明人计算:
	//z = (H1)η * (H2)ρ * mod Ñ
	var mx1 = ModPowInsecure(h1, eta, nTilde)
	var mx2 = ModPowInsecure(h2, rho, nTilde)
	var mx12 = new(big.Int).Mul(mx1, mx2)
	zkp.Z = new(big.Int).Mod(mx12, nTilde)
	if alpha.Sign() == -1 {
		alpha.Add(alpha, secp256k1.S256().P)
	}

	//u1 = (g)α in G
	zkp.U1 = PointMul(alpha, &Point{cx, cy})

	//u2 = (Γ)α * (β)N mod (N)2
	var my1 = ModPowInsecure(g, alpha, nSquared)
	var my2 = ModPowInsecure(beta, N, nSquared)
	var my12 = new(big.Int).Mul(my1, my2)
	zkp.U2 = new(big.Int).Mod(my12, nSquared)

	//u3 = (H1)α * (H2)γ mod N
	var mz1 = ModPowInsecure(h1, alpha, nTilde)
	var mz2 = ModPowInsecure(h2, gamma, nTilde)
	var mz12 = new(big.Int).Mul(mz1, mz2)
	zkp.U3 = new(big.Int).Mod(mz12, nTilde)

	//e = hash(g, y, w, z, u1 , u2 , u3)
	digest := Sha256Hash(GetBytes(g), Get2Bytes(cx, cy), GetBytes((w)),
		GetBytes((zkp.Z)), Get2Bytes(zkp.U1[0], zkp.U1[1]), GetBytes(zkp.U2), GetBytes(zkp.U3))
	if len(digest) == 0 {
		log.Crit("Assertion Error in zero knowledge proof when lock-in progress")
	}
	zkp.E = new(big.Int).SetBytes(digest)

	//s1 = eη + α
	var ee = new(big.Int).Mul(zkp.E, eta)
	zkp.S1 = new(big.Int).Add(ee, alpha)

	//s2 = (r)e *β mod N
	var re = ModPowInsecure(r, zkp.E, N) //k
	var rb = new(big.Int).Mul(re, beta)
	zkp.S2 = new(big.Int).Mod(rb, N)

	//S3 = eρ + γ
	var er = new(big.Int).Mul(zkp.E, rho)
	zkp.S3 = new(big.Int).Add(er, gamma)
}

func (zkp *Zkp) Verify(params *PublicParameters, rx, ry, w *big.Int) bool {
	var h1 = params.H1
	var h2 = params.H2
	var N = params.PaillierPubKey.N
	var nTilde = params.NTilde
	var nSquared = new(big.Int).Mul(N, N)
	var g = new(big.Int).Add(N, big.NewInt(1))
	var bitC = &ECPoint{
		X: secp256k1.S256().Gx,
		Y: secp256k1.S256().Gy,
	}
	valueCheckPassed := 4
	finishedU1 := make(chan bool, 1)
	finishedU2 := make(chan bool, 1)
	finishedU3 := make(chan bool, 1)
	finishedE := make(chan bool, 1)
	go zkp.checkU1(bitC.X, bitC.Y, rx, ry, nTilde, finishedU1)
	go zkp.checkU2(g, nSquared, N, w, finishedU2)
	go zkp.checkU3(h1, nTilde, h2, finishedU3)
	go zkp.checkE(bitC, w, g, finishedE)

	for {
		select {
		case checkU1 := <-finishedU1:
			if checkU1 == false {
				log.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(u1)")
				return false
			}
			log.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(u1)")
			valueCheckPassed--
		case checkU2 := <-finishedU2:
			if checkU2 == false {
				log.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(u2)")
				return false
			}
			log.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(u2)")
			valueCheckPassed--
		case checkV := <-finishedU3:
			if checkV == false {
				log.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(u3)")
				return false
			}
			log.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(u3)")
			valueCheckPassed--
		case checkE := <-finishedE:
			if checkE == false {
				log.Error("[LOCK-IN]Zero KnowLedge Proof failed when checking value(e)")
				return false
			}
			log.Info("[LOCK-IN]Zero KnowLedge Proof Success when checking value(e)")
			valueCheckPassed--
		}
		if valueCheckPassed == 0 {
			break
		}
	}
	return true
}

//checkU1 check:u1 = (g)s1 * (y)−e in G  (|g*s1 + (y)*−e -u1|=0,1 mean what?)
func (zkp *Zkp) checkU1(bx, by, rx, ry, nTilde *big.Int, notify chan bool) {
	g := &Point{bx, by}
	y := &Point{rx, ry}
	minuse := new(big.Int).Mul(zkp.E, big.NewInt(-1))
	minuse = new(big.Int).Mod(minuse, secp256k1.S256().N)
	u1 := pointAdd(PointMul(zkp.S1, g), PointMul(minuse, y))

	if u1[0].Cmp(zkp.U1[0]) == 0 && u1[1].Cmp(zkp.U1[1]) == 0 {
		notify <- true
		return
	} else {
		notify <- false
		return
	}
}

//checkU2 check:u2 = (Γ)s1 * (s2)N * (w)−e mod (N)2
func (zkp *Zkp) checkU2(g, nSquared, N, w *big.Int, notify chan bool) {
	var x = ModPowInsecure(g, zkp.S1, nSquared)
	var y = ModPowInsecure(zkp.S2, N, nSquared)
	var mulxy = new(big.Int).Mul(x, y)
	var c3neg = new(big.Int).Neg(zkp.E)
	var z = ModPowInsecure(w, c3neg, nSquared)
	var mulxyz = new(big.Int).Mul(mulxy, z)
	var result = new(big.Int).Mod(mulxyz, nSquared)

	if zkp.U2.Cmp(result) == 0 {
		notify <- true
		return
	} else {
		notify <- false
		return
	}
}

//checkU3 check:u3 = (H1)s1 *(H2)S3 * (z)−e mod Ñ
func (zkp *Zkp) checkU3(h1, nTilde, h2 *big.Int, notify chan bool) {
	var x = ModPowInsecure(h1, zkp.S1, nTilde)
	var y = ModPowInsecure(h2, zkp.S3, nTilde)
	var mulxy = new(big.Int).Mul(x, y)
	var eneg = new(big.Int).Neg(zkp.E)
	var z = ModPowInsecure(zkp.Z, eneg, nTilde)
	var mulxyz = new(big.Int).Mul(mulxy, z)
	var result = new(big.Int).Mod(mulxyz, nTilde)

	if zkp.U3.Cmp(result) == 0 {
		notify <- true
		return

	} else {
		notify <- false
		return
	}
}

//checkE check:e = hash(g,y,w,z,u1,u2,u3)
func (zkp *Zkp) checkE(bitC *ECPoint, w, g *big.Int, notify chan bool) {
	var result = Sha256Hash(GetBytes(g), Get2Bytes(bitC.X, bitC.Y), GetBytes(w),
		GetBytes(zkp.Z), Get2Bytes(zkp.U1[0], zkp.U1[1]), GetBytes(zkp.U2), GetBytes(zkp.U3))

	if zkp.E.Cmp(new(big.Int).SetBytes(result)) == 0 {
		notify <- true
		return
	} else {
		notify <- false
		return
	}
}
