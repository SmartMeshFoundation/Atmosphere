package proofs

import (
	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/secret_sharing"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm2/zkp"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

/// This is a proof of knowledge that a pair of group elements {D, E}
/// form a valid homomorphic ElGamal encryption (”in the exponent”) using public key Y .
/// (HEG is defined in B. Schoenmakers and P. Tuyls. Practical Two-Party Computation Based on the Conditional Gate)
/// Specifically, the witness is ω = (x, r), the statement is δ = (G, H, Y, D, E).
/// The relation R outputs 1 if D = xH+rY , E = rG (for the case of G=H this is ElGamal)
type HomoELGamalProof struct {
	T  *secret_sharing.GE
	A3 *secret_sharing.GE
	z1 *big.Int
	z2 *big.Int
}

type HomoElGamalWitness struct {
	r *big.Int
	x *big.Int
}

func NewHomoElGamalWitness(r, x *big.Int) *HomoElGamalWitness {
	return &HomoElGamalWitness{new(big.Int).Set(r), new(big.Int).Set(x)}
}

type HomoElGamalStatement struct {
	G *secret_sharing.GE
	H *secret_sharing.GE
	Y *secret_sharing.GE
	D *secret_sharing.GE
	E *secret_sharing.GE
}

func CreateHomoELGamalProof(w *HomoElGamalWitness, delta *HomoElGamalStatement) *HomoELGamalProof {
	s1 := zkp.RandomFromZn(secp256k1.S256().N)
	s2 := zkp.RandomFromZn(secp256k1.S256().N)

	A1x, A1y := S.ScalarMult(delta.H.X, delta.H.Y, s1.Bytes())
	A2x, A2y := S.ScalarMult(delta.Y.X, delta.Y.Y, s2.Bytes())
	A3x, A3y := S.ScalarMult(delta.G.X, delta.G.Y, s2.Bytes())
	tx, ty := secret_sharing.PointAdd(A1x, A1y, A2x, A2y)
	e := CreateHashFromGE([]*secret_sharing.GE{{tx, ty}, {A3x, A3y}, delta.G, delta.H, delta.Y, delta.D, delta.E})
	z1 := new(big.Int).Set(s1)
	if w.x.Cmp(big.NewInt(0)) != 0 {
		t := new(big.Int).Set(e)
		t = secret_sharing.ModMul(t, w.x)
		z1 = secret_sharing.ModAdd(z1, t)
	}
	t := new(big.Int).Set(e)
	t = secret_sharing.ModMul(t, w.r)
	z2 := new(big.Int).Set(s2)
	secret_sharing.ModAdd(z2, t)
	return &HomoELGamalProof{
		T:  &secret_sharing.GE{tx, ty},
		A3: &secret_sharing.GE{A3x, A3y},
		z1: z1,
		z2: z2,
	}

}

func (proof *HomoELGamalProof) Verify(delta *HomoElGamalStatement) bool {
	e := CreateHashFromGE([]*secret_sharing.GE{proof.T, proof.A3, delta.G, delta.H, delta.Y, delta.D, delta.E})
	//z12=z1*H+z2*Y
	z12x, z12y := S.ScalarMult(delta.H.X, delta.H.Y, proof.z1.Bytes())
	x, y := S.ScalarMult(delta.Y.X, delta.Y.Y, proof.z2.Bytes())
	z12x, z12y = secret_sharing.PointAdd(z12x, z12y, x, y)

	//T+e*D
	x, y = S.ScalarMult(delta.D.X, delta.D.Y, e.Bytes())
	tedx, tedy := secret_sharing.PointAdd(x, y, proof.T.X, proof.T.Y)
	//z2g=G*z2
	z2gx, z2gy := S.ScalarMult(delta.G.X, delta.G.Y, proof.z2.Bytes())

	//A3+e*E
	x, y = S.ScalarMult(delta.E.X, delta.E.Y, e.Bytes())
	a3eex, a3eey := secret_sharing.PointAdd(x, y, proof.A3.X, proof.A3.Y)

	if z12x.Cmp(tedx) == 0 && z12y.Cmp(tedy) == 0 &&
		z2gx.Cmp(a3eex) == 0 && z2gy.Cmp(a3eey) == 0 {
		return true
	}
	return false
}
func CreateHashFromGE(ge []*secret_sharing.GE) *big.Int {
	var bs [][]byte
	for _, g := range ge {
		bs = append(bs, g.X.Bytes())
		bs = append(bs, g.Y.Bytes())
	}
	hash := utils.Sha3(bs...)
	result := new(big.Int).SetBytes(hash[:])
	return secret_sharing.BigInt2PrivateKey(result)
}

/*func create_hash_from_ge(ge ...*ECPoint) *big.Int{
	var digest=sha256.New()
	for _,v:=range ge{

		tmp:=kgcenter.Get2Bytes(v.X,v.X)
		//digest=append(digest,tmp)
		digest.Write(tmp)
	}
	return new(big.Int).SetBytes(digest.Sum([]byte{}))
}*/

func pk_to_key_slice() {

}
