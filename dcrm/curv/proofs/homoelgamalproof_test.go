package proofs

import (
	"testing"

	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/secret_sharing"
)

func TestCreateHomoELGamalProof(t *testing.T) {
	witness := &HomoElGamalWitness{
		x: secret_sharing.RandomPrivateKey(),
		r: secret_sharing.RandomPrivateKey(),
	}

	h := secret_sharing.RandomPrivateKey()
	Hx, Hy := S.ScalarBaseMult(h.Bytes())
	y := secret_sharing.RandomPrivateKey()
	Yx, Yy := S.ScalarBaseMult(y.Bytes())
	tx := new(big.Int).Set(Hx)
	ty := new(big.Int).Set(Hy)
	tx, ty = S.ScalarMult(tx, ty, witness.x.Bytes())
	tx2, ty2 := S.ScalarMult(Yx, Yy, witness.r.Bytes())

	Dx, Dy := secret_sharing.PointAdd(tx, ty, tx2, ty2)

	Ex, Ey := S.ScalarBaseMult(witness.r.Bytes())
	delta := &HomoElGamalStatement{
		G: &secret_sharing.GE{S.Gx, S.Gy},
		H: &secret_sharing.GE{Hx, Hy},
		Y: &secret_sharing.GE{Yx, Yy},
		D: &secret_sharing.GE{Dx, Dy},
		E: &secret_sharing.GE{Ex, Ey},
	}

	prove := CreateHomoELGamalProof(witness, delta)
	if !prove.Verify(delta) {
		t.Error("not pass")
	}
}

/*func TestCreateHomoELGamalProof(t *testing.T) {
	witness:=&HomoElGamalWitness{
		kgcenter.RandomFromZn(secp256k1.S256().N),
		kgcenter.RandomFromZn(secp256k1.S256().N),
	}
	G:=&ECPoint{secp256k1.S256().Gx,secp256k1.S256().Gy}
	h:=kgcenter.RandomFromZn(secp256k1.S256().N)
	Hx,Hy :=secp256k1.S256().ScalarMult(G.X,G.Y,h.Bytes())
	y:=kgcenter.RandomFromZn(secp256k1.S256().N)
	Yx,Yy:=secp256k1.S256().ScalarMult(G.X,G.Y,y.Bytes())

	D:=secp256k1.S256().Add(&kgcenter.Point{G.X,G.Y},
		kgcenter.PointMul(witness.r,Y))
	E:=kgcenter.PointMul(witness.r,&kgcenter.Point{G.X,G.Y})
	delta:=&HomoElGamalStatement{
		G,
		&ECPoint{H[0],H[1]},
		&ECPoint{Y[0],Y[1]},
		&ECPoint{D[0],D[1]},
		&ECPoint{E[0],E[1]},
	}
	prove:=CreateHomoELGamalProof(witness,delta)
	if !prove.Verify(delta){
		t.Error("not pass")
	}
}*/
