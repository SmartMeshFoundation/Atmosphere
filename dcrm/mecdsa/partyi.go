package mecdsa

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"crypto/rand"
	mathrand "math/rand"
	"time"

	"errors"

	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/DistributedControlRightManagement/kgcenter/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/proofs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/secret_sharing"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
)

var SecureRnd = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
var masterPK = commitments.GenerateNMMasterPublicKey()

type Parameters struct {
	Threshold  int
	ShareCount int
}

type Keys struct {
	ui         *big.Int           //私钥片
	yi         *secret_sharing.GE //ui对应的公钥
	dk         *proofs.PrivateKey //paillier加密用
	partyIndex int                //自己是第几个
}

type KeyGenBroadcastMessage1 struct {
	e               *proofs.PublicKey         //paillier 公钥
	com             *big.Int                  //包含私钥片公钥信息的hash值
	correctKeyProof *proofs.NICorrectKeyProof //证明拥有一个paillier的私钥?
}

type SharedKeys struct {
	y  *secret_sharing.GE //公钥片之和,每个人拿到的都应该一样
	xi *big.Int           //其他人给我的vss秘密之和
}
type SignKeys struct {
	s       []int
	wi      *big.Int
	gwi     *secret_sharing.GE
	ki      *big.Int
	gammaI  *big.Int
	gGammaI *secret_sharing.GE
}
type SignBroadcastPhase1 struct {
	com *big.Int
}
type LocalSignature struct {
	li   *big.Int
	rhoi *big.Int
	R    *secret_sharing.GE
	si   *big.Int
	m    *big.Int
	y    *secret_sharing.GE
}

type Phase5Com1 struct {
	com *big.Int
}
type Phase5Com2 struct {
	com *big.Int
}
type Phase5ADecom1 struct {
	vi          *secret_sharing.GE
	ai          *secret_sharing.GE
	bi          *secret_sharing.GE
	blindFactor *big.Int
}
type Phase5DDecom2 struct {
	ui          *secret_sharing.GE
	ti          *secret_sharing.GE
	blindFactor *big.Int
}

//const
func createKeys(index int) *Keys {
	ui := secret_sharing.RandomPrivateKey()
	yi_x, yi_y := secp256k1.S256().ScalarBaseMult(ui.Bytes())
	yi := &secret_sharing.GE{yi_x, yi_y}
	partyPaillierPrivateKey, err := proofs.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("generate paillier key failed")
	}
	return &Keys{
		ui,
		yi,
		partyPaillierPrivateKey,
		index,
	}
}

//const
func CreateCommitmentWithUserDefinedRandomNess(message *big.Int, blindingFactor *big.Int) *big.Int {
	hash := utils.ShaSecret(message.Bytes(), blindingFactor.Bytes())
	b := new(big.Int)
	b.SetBytes(hash[:])
	return b
}

//const
func (k *Keys) phase1BroadcastPhase3ProofOfCorrectKey() (*KeyGenBroadcastMessage1, *big.Int) {
	blind_factor := secret_sharing.RandomPrivateKey()
	correctKeyProof := proofs.CreateNICorrectKeyProof(k.dk)
	com := CreateCommitmentWithUserDefinedRandomNess(k.yi.X, blind_factor)
	bcm1 := &KeyGenBroadcastMessage1{
		e:               k.dk.PublicKey.Clone(),
		com:             com,
		correctKeyProof: correctKeyProof,
	}
	return bcm1, blind_factor
}

/*
每个人都校验其他人送过来的私钥片的部分公钥,以及同态加密的公约,如果就通过vss机制分享自己的私钥片信息
私钥片是可以通过Recontruct恢复的.
const
*/
func (k *Keys) phase1VerifyComPhase3VerifyCorrectKeyPhase2Distribute(params *Parameters,
	blind_vec []*big.Int,
	y_vec []*secret_sharing.GE,
	bc1_vec []*KeyGenBroadcastMessage1) (vss *secret_sharing.VerifiableSS, secretShares []*big.Int, index int, err error) {
	///test length
	if len(blind_vec) != params.ShareCount {
		panic("随机数数量不对")
	}
	if len(bc1_vec) != params.ShareCount {
		panic("广播数量不对")
	}
	if len(y_vec) != params.ShareCount {
		panic("公钥数量不对")
	}

	for i := 0; i < len(bc1_vec); i++ {
		if CreateCommitmentWithUserDefinedRandomNess(y_vec[i].X, blind_vec[i]).Cmp(bc1_vec[i].com) == 0 && bc1_vec[i].correctKeyProof.Verify(bc1_vec[i].e) {
			continue
		}
		err = errors.New("invalid key")
		return
	}
	vss, secretShares = secret_sharing.Share(params.Threshold, params.ShareCount, k.ui)
	index = k.partyIndex
	return
}

/*
y_vec 私钥片对应公钥集合.todo 什么时候广播告诉所有公证人的呢?
const
*/
func (k *Keys) phase2_verify_vss_construct_keypair_phase3_pok_dlog(params *Parameters,
	y_vec []*secret_sharing.GE, secret_shares_vec []*big.Int, vss_scheme_vec []*secret_sharing.VerifiableSS, index int) (*SharedKeys, *proofs.DLogProof, error) {
	if len(y_vec) != params.ShareCount ||
		len(secret_shares_vec) != params.ShareCount ||
		len(vss_scheme_vec) != params.ShareCount {
		panic("arg error")
	}
	for i := 0; i < len(y_vec); i++ {
		//验证私钥片部分信息与公证人编号是一一对应关系
		if vss_scheme_vec[i].ValidateShare(secret_shares_vec[i], index) &&
			//私钥片对应的公钥就是多项式的常数项
			EqualGE(vss_scheme_vec[i].Commitments[0], y_vec[i]) {
			continue
		}
		return nil, nil, errors.New("invalid key")
	}
	//计算公钥片之和
	y0 := y_vec[0].Clone()
	for i := 1; i < len(y_vec); i++ {
		y0.X, y0.Y = secret_sharing.PointAdd(y0.X, y0.Y, y_vec[i].X, y_vec[i].Y)
	}
	x0 := big.NewInt(0)
	for i := 0; i < len(secret_shares_vec); i++ {
		secret_sharing.ModAdd(x0, secret_shares_vec[i])
	}
	dproof := proofs.Prove(x0)
	return &SharedKeys{y0, x0}, dproof, nil
}

//const
func VerifyDlogProofs(params *Parameters, dproofs []*proofs.DLogProof) bool {
	if len(dproofs) != params.ShareCount {
		panic("arg error")
	}
	for i := 0; i < params.ShareCount; i++ {
		if !proofs.Verify(dproofs[i]) {
			return false
		}
	}
	return true
}

//const
func createSignKeys(sharedKeys *SharedKeys, vss *secret_sharing.VerifiableSS, index int, s []int) *SignKeys {
	li := vss.MapShareToNewParams(index, s)
	wi := secret_sharing.ModMul(li, sharedKeys.xi)
	gwiX, gwiY := secret_sharing.S.ScalarBaseMult(wi.Bytes())
	gammaI := secret_sharing.RandomPrivateKey()
	gGammaIX, gGammaIY := secret_sharing.S.ScalarBaseMult(gammaI.Bytes())
	return &SignKeys{
		s:       s,
		wi:      wi,
		gwi:     &secret_sharing.GE{gwiX, gwiY},
		ki:      secret_sharing.RandomPrivateKey(),
		gammaI:  gammaI,
		gGammaI: &secret_sharing.GE{gGammaIX, gGammaIY},
	}
}
func (k *SignKeys) testPhase1Broadcast(blindFactor *big.Int) (*SignBroadcastPhase1, *big.Int) {
	gGammaIX, _ := secret_sharing.S.ScalarBaseMult(k.gammaI.Bytes())
	com := CreateCommitmentWithUserDefinedRandomNess(gGammaIX, blindFactor)
	return &SignBroadcastPhase1{com}, blindFactor
}

//const
func (k *SignKeys) phase1Broadcast() (*SignBroadcastPhase1, *big.Int) {
	blindFactor := secret_sharing.RandomPrivateKey()
	gGammaIX, _ := secret_sharing.S.ScalarBaseMult(k.gammaI.Bytes())
	com := CreateCommitmentWithUserDefinedRandomNess(gGammaIX, blindFactor)
	return &SignBroadcastPhase1{com}, blindFactor
}

//const
func (k *SignKeys) phase2DeltaI(alpha_vec []*big.Int, beta_vec []*big.Int) *big.Int {
	if len(alpha_vec) != len(beta_vec) ||
		len(alpha_vec) != len(k.s)-1 {
		panic("arg error")
	}
	kiGammaI := new(big.Int).Set(k.ki)
	secret_sharing.ModMul(kiGammaI, k.gammaI)
	for i := 0; i < len(alpha_vec); i++ {
		a := new(big.Int).Set(alpha_vec[i])
		secret_sharing.ModAdd(a, beta_vec[i])
		secret_sharing.ModAdd(kiGammaI, a)
	}
	return kiGammaI
}

//const
func (k *SignKeys) phase2SigmaI(miu_vec []*big.Int, ni_vec []*big.Int) *big.Int {
	if len(miu_vec) != len(ni_vec) ||
		len(miu_vec) != len(k.s)-1 {
		panic("length error")
	}
	kiwi := new(big.Int).Set(k.ki)
	secret_sharing.ModMul(kiwi, k.wi)
	for i := 0; i < len(miu_vec); i++ {
		secret_sharing.ModAdd(kiwi, miu_vec[i])
		secret_sharing.ModAdd(kiwi, ni_vec[i])
	}
	return kiwi
}

//const
func phase3ReconstructDelta(delta_vec []*big.Int) *big.Int {
	sum := big.NewInt(0)
	for i := 0; i < len(delta_vec); i++ {
		secret_sharing.ModAdd(sum, delta_vec[i])
	}
	return secret_sharing.Invert(sum, secret_sharing.S.N)
}

//const
func phase4(delta_inv *big.Int,
	b_proof_vec []*proofs.DLogProof,
	blind_vec []*big.Int,
	g_gamma_i_vec []*secret_sharing.GE,
	bc1_vec []*SignBroadcastPhase1) (*secret_sharing.GE, error) {
	for i := 0; i < len(b_proof_vec); i++ {
		if EqualGE(b_proof_vec[i].PK, g_gamma_i_vec[i]) &&
			CreateCommitmentWithUserDefinedRandomNess(g_gamma_i_vec[i].X, blind_vec[i]).Cmp(bc1_vec[i].com) == 0 {
			continue
		}
		return nil, errors.New("invliad key")
	}
	gc := g_gamma_i_vec[0].Clone()
	sumx, sumy := gc.X, gc.Y
	for i := 1; i < len(g_gamma_i_vec); i++ {
		sumx, sumy = secret_sharing.PointAdd(sumx, sumy, g_gamma_i_vec[i].X, g_gamma_i_vec[i].Y)
	}
	rx, ry := secret_sharing.S.ScalarMult(sumx, sumy, delta_inv.Bytes())
	return &secret_sharing.GE{rx, ry}, nil
}

//const
func phase5LocalSignature(ki *big.Int, message *big.Int,
	R *secret_sharing.GE, sigmaI *big.Int,
	pubkey *secret_sharing.GE) *LocalSignature {
	m := secret_sharing.BigInt2PrivateKey(message)
	r := new(big.Int).Set(R.X)
	r = secret_sharing.BigInt2PrivateKey(r.Mod(r, secret_sharing.S.N))
	si := secret_sharing.ModMul(m, ki)
	secret_sharing.ModMul(r, sigmaI)
	secret_sharing.ModAdd(si, r) //si=m * k_i + r * sigma_i
	return &LocalSignature{
		li:   secret_sharing.RandomPrivateKey(), 
		rhoi: secret_sharing.RandomPrivateKey(),
		//li:   big.NewInt(71),
		//rhoi: big.NewInt(73),
		R: &secret_sharing.GE{
			X: new(big.Int).Set(R.X),
			Y: new(big.Int).Set(R.Y),
		},
		si: si,
		m:  new(big.Int).Set(message),
		y: &secret_sharing.GE{
			X: new(big.Int).Set(pubkey.X),
			Y: new(big.Int).Set(pubkey.Y),
		},
	}

}
func (l *LocalSignature) testphase5aBroadcast5bZkproof(blindFactor *big.Int) (*Phase5Com1, *Phase5ADecom1, *proofs.HomoELGamalProof) {
	aix, aiy := secret_sharing.S.ScalarBaseMult(l.rhoi.Bytes())
	l_i_rho_i := new(big.Int).Set(l.li)
	secret_sharing.ModMul(l_i_rho_i, l.rhoi)
	//G*l_i_rho_i
	bix, biy := secret_sharing.S.ScalarBaseMult(l_i_rho_i.Bytes())
	//vi=R*si+G*li
	tx, ty := secret_sharing.S.ScalarMult(l.R.X, l.R.Y, l.si.Bytes())
	vix, viy := secret_sharing.S.ScalarBaseMult(l.li.Bytes())
	vix, viy = secret_sharing.PointAdd(vix, viy, tx, ty)

	inputhash := proofs.CreateHashFromGE([]*secret_sharing.GE{
		{vix, viy}, {aix, aiy}, {bix, biy},
	})
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash, blindFactor)

	witness := proofs.NewHomoElGamalWitness(l.li, l.si)
	delta := &proofs.HomoElGamalStatement{
		G: secret_sharing.NewGE(aix, aiy),
		H: secret_sharing.NewGE(l.R.X, l.R.Y),
		Y: secret_sharing.NewGE(secret_sharing.S.Gx, secret_sharing.S.Gy),
		D: secret_sharing.NewGE(vix, viy),
		E: secret_sharing.NewGE(bix, biy),
	}
	proof := proofs.CreateHomoELGamalProof(witness, delta)
	return &Phase5Com1{com},
		&Phase5ADecom1{
			vi:          secret_sharing.NewGE(vix, viy),
			ai:          secret_sharing.NewGE(aix, aiy),
			bi:          secret_sharing.NewGE(bix, biy),
			blindFactor: blindFactor,
		},
		proof
}

//const
func (l *LocalSignature) phase5aBroadcast5bZkproof() (*Phase5Com1, *Phase5ADecom1, *proofs.HomoELGamalProof) {
	blindFactor := secret_sharing.RandomPrivateKey()
	aix, aiy := secret_sharing.S.ScalarBaseMult(l.rhoi.Bytes())
	l_i_rho_i := new(big.Int).Set(l.li)
	secret_sharing.ModMul(l_i_rho_i, l.rhoi)
	//G*l_i_rho_i
	bix, biy := secret_sharing.S.ScalarBaseMult(l_i_rho_i.Bytes())
	//vi=R*si+G*li
	tx, ty := secret_sharing.S.ScalarMult(l.R.X, l.R.Y, l.si.Bytes())
	vix, viy := secret_sharing.S.ScalarBaseMult(l.li.Bytes())
	vix, viy = secret_sharing.PointAdd(vix, viy, tx, ty)

	inputhash := proofs.CreateHashFromGE([]*secret_sharing.GE{
		{vix, viy}, {aix, aiy}, {bix, biy},
	})
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash, blindFactor)

	witness := proofs.NewHomoElGamalWitness(l.li, l.si)
	delta := &proofs.HomoElGamalStatement{
		G: secret_sharing.NewGE(aix, aiy),
		H: secret_sharing.NewGE(l.R.X, l.R.Y),
		Y: secret_sharing.NewGE(secret_sharing.S.Gx, secret_sharing.S.Gy),
		D: secret_sharing.NewGE(vix, viy),
		E: secret_sharing.NewGE(bix, biy),
	}
	proof := proofs.CreateHomoELGamalProof(witness, delta)
	return &Phase5Com1{com},
		&Phase5ADecom1{
			vi:          secret_sharing.NewGE(vix, viy),
			ai:          secret_sharing.NewGE(aix, aiy),
			bi:          secret_sharing.NewGE(bix, biy),
			blindFactor: blindFactor,
		},
		proof
}

//const
func (l *LocalSignature) phase5c(decomVec []*Phase5ADecom1, comVec []*Phase5Com1,
	elgamalProofs []*proofs.HomoELGamalProof,
	vi *secret_sharing.GE,
	R *secret_sharing.GE,
) (*Phase5Com2, *Phase5DDecom2, error) {
	if len(decomVec) != len(comVec) {
		panic("arg error")
	}
	g := secret_sharing.NewGE(secret_sharing.S.Gx, secret_sharing.S.Gy)
	for i := 0; i < len(comVec); i++ {
		delta := &proofs.HomoElGamalStatement{
			G: decomVec[i].ai,
			H: R,
			Y: g,
			D: decomVec[i].vi,
			E: decomVec[i].bi,
		}
		inputhash := proofs.CreateHashFromGE([]*secret_sharing.GE{
			decomVec[i].vi,
			decomVec[i].ai,
			decomVec[i].bi,
		})
		e := CreateCommitmentWithUserDefinedRandomNess(inputhash, decomVec[i].blindFactor)
		if e.Cmp(comVec[i].com) == 0 &&
			elgamalProofs[i].Verify(delta) {
			continue
		}
		return nil, nil, errors.New("invalid com")
	}
	v := vi.Clone()
	for i := 0; i < len(comVec); i++ {
		v.X, v.Y = secret_sharing.PointAdd(v.X, v.Y, decomVec[i].vi.X, decomVec[i].vi.Y)
	}
	a := decomVec[0].ai.Clone()
	for i := 1; i < len(comVec); i++ {
		a.X, a.Y = secret_sharing.PointAdd(a.X, a.Y, decomVec[i].ai.X, decomVec[i].ai.Y)
	}
	r := secret_sharing.BigInt2PrivateKey(l.R.X)
	yrx, yry := secret_sharing.S.ScalarMult(l.y.X, l.y.Y, r.Bytes())
	m := secret_sharing.BigInt2PrivateKey(l.m)
	gmx, gmy := secret_sharing.S.ScalarBaseMult(m.Bytes())
	v.X, v.Y = secret_sharing.PointSub(v.X, v.Y, gmx, gmy)
	v.X, v.Y = secret_sharing.PointSub(v.X, v.Y, yrx, yry)

	uix, uiy := secret_sharing.S.ScalarMult(v.X, v.Y, l.rhoi.Bytes())
	tix, tiy := secret_sharing.S.ScalarMult(a.X, a.Y, l.li.Bytes())

	inputhash := proofs.CreateHashFromGE([]*secret_sharing.GE{
		{uix, uiy},
		{tix, tiy},
	})
	blindFactor := secret_sharing.RandomPrivateKey()
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash, blindFactor)
	return &Phase5Com2{com},
		&Phase5DDecom2{
			ui:          &secret_sharing.GE{uix, uiy},
			ti:          &secret_sharing.GE{tix, tiy},
			blindFactor: blindFactor,
		},
		nil
}

func (l *LocalSignature) testphase5c(decomVec []*Phase5ADecom1, comVec []*Phase5Com1,
	elgamalProofs []*proofs.HomoELGamalProof,
	vi *secret_sharing.GE,
	R *secret_sharing.GE, blindFactor *big.Int,
) (*Phase5Com2, *Phase5DDecom2, error) {
	if len(decomVec) != len(comVec) {
		panic("arg error")
	}
	g := secret_sharing.NewGE(secret_sharing.S.Gx, secret_sharing.S.Gy)
	for i := 0; i < len(comVec); i++ {
		delta := &proofs.HomoElGamalStatement{
			G: decomVec[i].ai,
			H: R,
			Y: g,
			D: decomVec[i].vi,
			E: decomVec[i].bi,
		}
		inputhash := proofs.CreateHashFromGE([]*secret_sharing.GE{
			decomVec[i].vi,
			decomVec[i].ai,
			decomVec[i].bi,
		})
		e := CreateCommitmentWithUserDefinedRandomNess(inputhash, decomVec[i].blindFactor)
		if e.Cmp(comVec[i].com) == 0 &&
			elgamalProofs[i].Verify(delta) {
			continue
		}
		return nil, nil, errors.New("invalid com")
	}
	v := vi.Clone()
	for i := 0; i < len(comVec); i++ {
		v.X, v.Y = secret_sharing.PointAdd(v.X, v.Y, decomVec[i].vi.X, decomVec[i].vi.Y)
	}
	a := decomVec[0].ai.Clone()
	for i := 1; i < len(comVec); i++ {
		a.X, a.Y = secret_sharing.PointAdd(a.X, a.Y, decomVec[i].ai.X, decomVec[i].ai.Y)
	}
	r := secret_sharing.BigInt2PrivateKey(l.R.X)
	yrx, yry := secret_sharing.S.ScalarMult(l.y.X, l.y.Y, r.Bytes())
	m := secret_sharing.BigInt2PrivateKey(l.m)
	gmx, gmy := secret_sharing.S.ScalarBaseMult(m.Bytes())
	v.X, v.Y = secret_sharing.PointSub(v.X, v.Y, gmx, gmy)
	v.X, v.Y = secret_sharing.PointSub(v.X, v.Y, yrx, yry)

	uix, uiy := secret_sharing.S.ScalarMult(v.X, v.Y, l.rhoi.Bytes())
	tix, tiy := secret_sharing.S.ScalarMult(a.X, a.Y, l.li.Bytes())

	inputhash := proofs.CreateHashFromGE([]*secret_sharing.GE{
		{uix, uiy},
		{tix, tiy},
	})
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash, blindFactor)
	return &Phase5Com2{com},
		&Phase5DDecom2{
			ui:          &secret_sharing.GE{uix, uiy},
			ti:          &secret_sharing.GE{tix, tiy},
			blindFactor: blindFactor,
		},
		nil
}

//const
func (l *LocalSignature) phase5d(decom_vec2 []*Phase5DDecom2,
	com_vec2 []*Phase5Com2, decom_vec1 []*Phase5ADecom1) (*big.Int, error) {
	if len(decom_vec1) != len(decom_vec2) ||
		len(decom_vec2) != len(com_vec2) {
		panic("arg error")
	}
	for i := 0; i < len(com_vec2); i++ {
		inputhash := proofs.CreateHashFromGE([]*secret_sharing.GE{decom_vec2[i].ui, decom_vec2[i].ti})
		inputhash = CreateCommitmentWithUserDefinedRandomNess(inputhash, decom_vec2[i].blindFactor)
		if inputhash.Cmp(com_vec2[i].com) != 0 {
			return nil, errors.New("invalid com")
		}
	}

	biased_sum_tbX := new(big.Int).Set(secret_sharing.S.Gx)
	biased_sum_tbY := new(big.Int).Set(secret_sharing.S.Gy)

	for i := 0; i < len(com_vec2); i++ {
		biased_sum_tbX, biased_sum_tbY = secret_sharing.PointAdd(biased_sum_tbX, biased_sum_tbY,
			decom_vec2[i].ti.X, decom_vec2[i].ti.Y)
		biased_sum_tbX, biased_sum_tbY = secret_sharing.PointAdd(biased_sum_tbX, biased_sum_tbY,
			decom_vec1[i].bi.X, decom_vec1[i].bi.Y)
	}
	for i := 0; i < len(com_vec2); i++ {
		biased_sum_tbX, biased_sum_tbY = secret_sharing.PointSub(
			biased_sum_tbX, biased_sum_tbY,
			decom_vec2[i].ui.X, decom_vec2[i].ui.Y,
		)
	}
	log.Trace(fmt.Sprintf("(gx,gy)=(%s,%s)", secret_sharing.S.Gx.Text(16), secret_sharing.S.Gy.Text(16)))
	log.Trace(fmt.Sprintf("(tbx,tby)=(%s,%s)", biased_sum_tbX.Text(16), biased_sum_tbY.Text(16)))
	if secret_sharing.S.Gx.Cmp(biased_sum_tbX) == 0 &&
		secret_sharing.S.Gy.Cmp(biased_sum_tbY) == 0 {
		return new(big.Int).Set(l.si), nil
	}
	return nil, errors.New("invalid key")
}

//const
func (l *LocalSignature) outputSignature(s_vec []*big.Int) (r, s *big.Int, err error) {
	s = new(big.Int).Set(l.si)
	for i := 0; i < len(s_vec); i++ {
		secret_sharing.ModAdd(s, s_vec[i])
	}
	r = secret_sharing.BigInt2PrivateKey(l.R.X)
	if !verify(s, r, l.y, l.m) {
		err = errors.New("invilad signature")
	}
	return
}

//const
func verify(s, r *big.Int, y *secret_sharing.GE, message *big.Int) bool {
	b := secret_sharing.Invert(s, secret_sharing.S.N)
	a := secret_sharing.BigInt2PrivateKey(message)
	u1 := new(big.Int).Set(a)
	u1 = secret_sharing.ModMul(u1, b)
	u2 := new(big.Int).Set(r)
	u2 = secret_sharing.ModMul(u2, b)

	gu1x, gu1y := secret_sharing.S.ScalarBaseMult(u1.Bytes())
	yu2x, yu2y := secret_sharing.S.ScalarMult(y.X, y.Y, u2.Bytes())
	gu1x, gu1y = secret_sharing.PointAdd(gu1x, gu1y, yu2x, yu2y)
	if secret_sharing.BigInt2PrivateKey(gu1x).Cmp(r) == 0 {
		return true
	}
	return false

}
