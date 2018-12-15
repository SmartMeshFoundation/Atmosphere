package mecdsa

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"crypto/rand"
	mathrand "math/rand"
	"time"

	"errors"

	"fmt"

	"bytes"

	"github.com/SmartMeshFoundation/Atmosphere/DistributedControlRightManagement/kgcenter/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/feldman"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/proofs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/share"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var SecureRnd = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
var masterPK = commitments.GenerateNMMasterPublicKey()

type Parameters struct {
	Threshold  int
	ShareCount int
}

type Keys struct {
	ui         share.SPrivKey     //原始随机数
	yi         *share.SPubKey     //ui对应的公钥
	dk         *proofs.PrivateKey //paillier加密用
	partyIndex int                //自己是第几个
}

type KeyGenBroadcastMessage1 struct {
	e               *proofs.PublicKey         //paillier 公钥
	com             *big.Int                  //包含私钥片公钥信息的hash值
	correctKeyProof *proofs.NICorrectKeyProof //证明拥有一个paillier的私钥?
}

type SharedKeys struct {
	y  *share.SPubKey //公钥片之和,每个人拿到的都应该一样
	xi share.SPrivKey //其他人给我的vss秘密之和,也就是我的真正私钥片
}
type SignKeys struct {
	s       []int
	wi      share.SPrivKey
	gwi     *share.SPubKey
	ki      share.SPrivKey
	gammaI  share.SPrivKey
	gGammaI *share.SPubKey
}
type SignBroadcastPhase1 struct {
	com *big.Int
}
type LocalSignature struct {
	li   share.SPrivKey
	rhoi share.SPrivKey
	R    *share.SPubKey
	si   share.SPrivKey
	m    *big.Int
	y    *share.SPubKey
}

type Phase5Com1 struct {
	com *big.Int
}
type Phase5Com2 struct {
	com *big.Int
}
type Phase5ADecom1 struct {
	vi          *share.SPubKey
	ai          *share.SPubKey
	bi          *share.SPubKey
	blindFactor *big.Int
}
type Phase5DDecom2 struct {
	ui          *share.SPubKey
	ti          *share.SPubKey
	blindFactor *big.Int
}

//const
func createKeys(index int) *Keys {
	ui := share.RandomPrivateKey()
	yi_x, yi_y := secp256k1.S256().ScalarBaseMult(ui.Bytes())
	yi := &share.SPubKey{yi_x, yi_y}
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
	blind_factor := share.RandomBigInt()
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
	y_vec []*share.SPubKey,
	bc1_vec []*KeyGenBroadcastMessage1) (vss *feldman.VerifiableSS, secretShares []share.SPrivKey, index int, err error) {
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
	vss, secretShares = feldman.Share(params.Threshold, params.ShareCount, k.ui)
	index = k.partyIndex
	return
}

/*
y_vec 私钥片对应公钥集合.todo 什么时候广播告诉所有公证人的呢?
const
*/
func (k *Keys) phase2_verify_vss_construct_keypair_phase3_pok_dlog(params *Parameters,
	y_vec []*share.SPubKey, secret_shares_vec []share.SPrivKey, vss_scheme_vec []*feldman.VerifiableSS, index int) (*SharedKeys, *proofs.DLogProof, error) {
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
		y0.X, y0.Y = share.PointAdd(y0.X, y0.Y, y_vec[i].X, y_vec[i].Y)
	}
	x0 := share.PrivKeyZero.Clone()
	for i := 0; i < len(secret_shares_vec); i++ {
		share.ModAdd(x0, secret_shares_vec[i])
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
func createSignKeys(sharedKeys *SharedKeys, vss *feldman.VerifiableSS, index int, s []int) *SignKeys {
	li := vss.MapShareToNewParams(index, s)
	wi := share.ModMul(li, sharedKeys.xi)
	gwiX, gwiY := share.S.ScalarBaseMult(wi.Bytes())
	gammaI := share.RandomPrivateKey()
	gGammaIX, gGammaIY := share.S.ScalarBaseMult(gammaI.Bytes())
	return &SignKeys{
		s:       s,
		wi:      wi,
		gwi:     &share.SPubKey{gwiX, gwiY},
		ki:      share.RandomPrivateKey(),
		gammaI:  gammaI,
		gGammaI: &share.SPubKey{gGammaIX, gGammaIY},
	}
}

//const
func (k *SignKeys) phase1Broadcast() (*SignBroadcastPhase1, *big.Int) {
	blindFactor := share.RandomBigInt()
	gGammaIX, _ := share.S.ScalarBaseMult(k.gammaI.Bytes())
	com := CreateCommitmentWithUserDefinedRandomNess(gGammaIX, blindFactor)
	return &SignBroadcastPhase1{com}, blindFactor
}

//const
func (k *SignKeys) phase2DeltaI(alpha_vec []share.SPrivKey, beta_vec []share.SPrivKey) share.SPrivKey {
	if len(alpha_vec) != len(beta_vec) ||
		len(alpha_vec) != len(k.s)-1 {
		panic("arg error")
	}
	kiGammaI := k.ki.Clone()
	share.ModMul(kiGammaI, k.gammaI)
	for i := 0; i < len(alpha_vec); i++ {
		share.ModAdd(kiGammaI, alpha_vec[i])
		share.ModAdd(kiGammaI, beta_vec[i])
	}
	return kiGammaI
}

//const
func (k *SignKeys) phase2SigmaI(miu_vec, ni_vec []share.SPrivKey) share.SPrivKey {
	if len(miu_vec) != len(ni_vec) ||
		len(miu_vec) != len(k.s)-1 {
		panic("length error")
	}
	kiwi := k.ki.Clone()
	share.ModMul(kiwi, k.wi)
	for i := 0; i < len(miu_vec); i++ {
		share.ModAdd(kiwi, miu_vec[i])
		share.ModAdd(kiwi, ni_vec[i])
	}
	return kiwi
}

//const
func phase3ReconstructDelta(delta_vec []share.SPrivKey) share.SPrivKey {
	sum := share.PrivKeyZero.Clone()
	for i := 0; i < len(delta_vec); i++ {
		share.ModAdd(sum, delta_vec[i])
	}
	return share.InvertN(sum)
}

//const
func phase4(delta_inv share.SPrivKey,
	b_proof_vec []*proofs.DLogProof,
	blind_vec []*big.Int,
	g_gamma_i_vec []*share.SPubKey,
	bc1_vec []*SignBroadcastPhase1) (*share.SPubKey, error) {
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
		sumx, sumy = share.PointAdd(sumx, sumy, g_gamma_i_vec[i].X, g_gamma_i_vec[i].Y)
	}
	rx, ry := share.S.ScalarMult(sumx, sumy, delta_inv.Bytes())
	return &share.SPubKey{rx, ry}, nil
}

//const
func phase5LocalSignature(ki share.SPrivKey, message *big.Int,
	R *share.SPubKey, sigmaI share.SPrivKey,
	pubkey *share.SPubKey) *LocalSignature {
	m := share.BigInt2PrivateKey(message)
	r := share.BigInt2PrivateKey(R.X)
	si := share.ModMul(m, ki)
	share.ModMul(r, sigmaI)
	share.ModAdd(si, r) //si=m * k_i + r * sigma_i
	return &LocalSignature{
		li:   share.RandomPrivateKey(),
		rhoi: share.RandomPrivateKey(),
		//li:   big.NewInt(71),
		//rhoi: big.NewInt(73),
		R: &share.SPubKey{
			X: new(big.Int).Set(R.X),
			Y: new(big.Int).Set(R.Y),
		},
		si: si,
		m:  new(big.Int).Set(message),
		y: &share.SPubKey{
			X: new(big.Int).Set(pubkey.X),
			Y: new(big.Int).Set(pubkey.Y),
		},
	}

}
func (l *LocalSignature) testphase5aBroadcast5bZkproof(blindFactor *big.Int) (*Phase5Com1, *Phase5ADecom1, *proofs.HomoELGamalProof) {
	aix, aiy := share.S.ScalarBaseMult(l.rhoi.Bytes())
	l_i_rho_i := l.li.Clone()
	share.ModMul(l_i_rho_i, l.rhoi)
	//G*l_i_rho_i
	bix, biy := share.S.ScalarBaseMult(l_i_rho_i.Bytes())
	//vi=R*si+G*li
	tx, ty := share.S.ScalarMult(l.R.X, l.R.Y, l.si.Bytes())
	vix, viy := share.S.ScalarBaseMult(l.li.Bytes())
	vix, viy = share.PointAdd(vix, viy, tx, ty)

	inputhash := proofs.CreateHashFromGE([]*share.SPubKey{
		{vix, viy}, {aix, aiy}, {bix, biy},
	})
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash.D, blindFactor)

	witness := proofs.NewHomoElGamalWitness(l.li, l.si)
	delta := &proofs.HomoElGamalStatement{
		G: share.NewGE(aix, aiy),
		H: share.NewGE(l.R.X, l.R.Y),
		Y: share.NewGE(share.S.Gx, share.S.Gy),
		D: share.NewGE(vix, viy),
		E: share.NewGE(bix, biy),
	}
	proof := proofs.CreateHomoELGamalProof(witness, delta)
	return &Phase5Com1{com},
		&Phase5ADecom1{
			vi:          share.NewGE(vix, viy),
			ai:          share.NewGE(aix, aiy),
			bi:          share.NewGE(bix, biy),
			blindFactor: blindFactor,
		},
		proof
}

//const
func (l *LocalSignature) phase5aBroadcast5bZkproof() (*Phase5Com1, *Phase5ADecom1, *proofs.HomoELGamalProof) {
	blindFactor := share.RandomBigInt()
	aix, aiy := share.S.ScalarBaseMult(l.rhoi.Bytes())
	l_i_rho_i := l.li.Clone()
	share.ModMul(l_i_rho_i, l.rhoi)
	//G*l_i_rho_i
	bix, biy := share.S.ScalarBaseMult(l_i_rho_i.Bytes())
	//vi=R*si+G*li
	tx, ty := share.S.ScalarMult(l.R.X, l.R.Y, l.si.Bytes())
	vix, viy := share.S.ScalarBaseMult(l.li.Bytes())
	vix, viy = share.PointAdd(vix, viy, tx, ty)

	inputhash := proofs.CreateHashFromGE([]*share.SPubKey{
		{vix, viy}, {aix, aiy}, {bix, biy},
	})
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash.D, blindFactor)

	witness := proofs.NewHomoElGamalWitness(l.li, l.si)
	delta := &proofs.HomoElGamalStatement{
		G: share.NewGE(aix, aiy),
		H: share.NewGE(l.R.X, l.R.Y),
		Y: share.NewGE(share.S.Gx, share.S.Gy),
		D: share.NewGE(vix, viy),
		E: share.NewGE(bix, biy),
	}
	proof := proofs.CreateHomoELGamalProof(witness, delta)
	return &Phase5Com1{com},
		&Phase5ADecom1{
			vi:          share.NewGE(vix, viy),
			ai:          share.NewGE(aix, aiy),
			bi:          share.NewGE(bix, biy),
			blindFactor: blindFactor,
		},
		proof
}

//const
func (l *LocalSignature) phase5c(decomVec []*Phase5ADecom1, comVec []*Phase5Com1,
	elgamalProofs []*proofs.HomoELGamalProof,
	vi *share.SPubKey,
	R *share.SPubKey,
) (*Phase5Com2, *Phase5DDecom2, error) {
	if len(decomVec) != len(comVec) {
		panic("arg error")
	}
	g := share.NewGE(share.S.Gx, share.S.Gy)
	for i := 0; i < len(comVec); i++ {
		delta := &proofs.HomoElGamalStatement{
			G: decomVec[i].ai,
			H: R,
			Y: g,
			D: decomVec[i].vi,
			E: decomVec[i].bi,
		}
		inputhash := proofs.CreateHashFromGE([]*share.SPubKey{
			decomVec[i].vi,
			decomVec[i].ai,
			decomVec[i].bi,
		})
		e := CreateCommitmentWithUserDefinedRandomNess(inputhash.D, decomVec[i].blindFactor)
		if e.Cmp(comVec[i].com) == 0 &&
			elgamalProofs[i].Verify(delta) {
			continue
		}
		return nil, nil, errors.New("invalid com")
	}
	v := vi.Clone()
	for i := 0; i < len(comVec); i++ {
		v.X, v.Y = share.PointAdd(v.X, v.Y, decomVec[i].vi.X, decomVec[i].vi.Y)
	}
	a := decomVec[0].ai.Clone()
	for i := 1; i < len(comVec); i++ {
		a.X, a.Y = share.PointAdd(a.X, a.Y, decomVec[i].ai.X, decomVec[i].ai.Y)
	}
	r := share.BigInt2PrivateKey(l.R.X)
	yrx, yry := share.S.ScalarMult(l.y.X, l.y.Y, r.Bytes())
	m := share.BigInt2PrivateKey(l.m)
	gmx, gmy := share.S.ScalarBaseMult(m.Bytes())
	v.X, v.Y = share.PointSub(v.X, v.Y, gmx, gmy)
	v.X, v.Y = share.PointSub(v.X, v.Y, yrx, yry)

	uix, uiy := share.S.ScalarMult(v.X, v.Y, l.rhoi.Bytes())
	tix, tiy := share.S.ScalarMult(a.X, a.Y, l.li.Bytes())

	inputhash := proofs.CreateHashFromGE([]*share.SPubKey{
		{uix, uiy},
		{tix, tiy},
	})
	blindFactor := share.RandomBigInt()
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash.D, blindFactor)
	return &Phase5Com2{com},
		&Phase5DDecom2{
			ui:          &share.SPubKey{uix, uiy},
			ti:          &share.SPubKey{tix, tiy},
			blindFactor: blindFactor,
		},
		nil
}

func (l *LocalSignature) testphase5c(decomVec []*Phase5ADecom1, comVec []*Phase5Com1,
	elgamalProofs []*proofs.HomoELGamalProof,
	vi *share.SPubKey,
	R *share.SPubKey, blindFactor *big.Int,
) (*Phase5Com2, *Phase5DDecom2, error) {
	if len(decomVec) != len(comVec) {
		panic("arg error")
	}
	g := share.NewGE(share.S.Gx, share.S.Gy)
	for i := 0; i < len(comVec); i++ {
		delta := &proofs.HomoElGamalStatement{
			G: decomVec[i].ai,
			H: R,
			Y: g,
			D: decomVec[i].vi,
			E: decomVec[i].bi,
		}
		inputhash := proofs.CreateHashFromGE([]*share.SPubKey{
			decomVec[i].vi,
			decomVec[i].ai,
			decomVec[i].bi,
		})
		e := CreateCommitmentWithUserDefinedRandomNess(inputhash.D, decomVec[i].blindFactor)
		if e.Cmp(comVec[i].com) == 0 &&
			elgamalProofs[i].Verify(delta) {
			continue
		}
		return nil, nil, errors.New("invalid com")
	}
	v := vi.Clone()
	for i := 0; i < len(comVec); i++ {
		v.X, v.Y = share.PointAdd(v.X, v.Y, decomVec[i].vi.X, decomVec[i].vi.Y)
	}
	a := decomVec[0].ai.Clone()
	for i := 1; i < len(comVec); i++ {
		a.X, a.Y = share.PointAdd(a.X, a.Y, decomVec[i].ai.X, decomVec[i].ai.Y)
	}
	r := share.BigInt2PrivateKey(l.R.X)
	yrx, yry := share.S.ScalarMult(l.y.X, l.y.Y, r.Bytes())
	m := share.BigInt2PrivateKey(l.m)
	gmx, gmy := share.S.ScalarBaseMult(m.Bytes())
	v.X, v.Y = share.PointSub(v.X, v.Y, gmx, gmy)
	v.X, v.Y = share.PointSub(v.X, v.Y, yrx, yry)

	uix, uiy := share.S.ScalarMult(v.X, v.Y, l.rhoi.Bytes())
	tix, tiy := share.S.ScalarMult(a.X, a.Y, l.li.Bytes())

	inputhash := proofs.CreateHashFromGE([]*share.SPubKey{
		{uix, uiy},
		{tix, tiy},
	})
	com := CreateCommitmentWithUserDefinedRandomNess(inputhash.D, blindFactor)
	return &Phase5Com2{com},
		&Phase5DDecom2{
			ui:          &share.SPubKey{uix, uiy},
			ti:          &share.SPubKey{tix, tiy},
			blindFactor: blindFactor,
		},
		nil
}

//const
func (l *LocalSignature) phase5d(decom_vec2 []*Phase5DDecom2,
	com_vec2 []*Phase5Com2, decom_vec1 []*Phase5ADecom1) (share.SPrivKey, error) {
	if len(decom_vec1) != len(decom_vec2) ||
		len(decom_vec2) != len(com_vec2) {
		panic("arg error")
	}
	for i := 0; i < len(com_vec2); i++ {
		inputhash := proofs.CreateHashFromGE([]*share.SPubKey{decom_vec2[i].ui, decom_vec2[i].ti})
		inputhash.D = CreateCommitmentWithUserDefinedRandomNess(inputhash.D, decom_vec2[i].blindFactor)
		if inputhash.D.Cmp(com_vec2[i].com) != 0 {
			return share.PrivKeyZero, errors.New("invalid com")
		}
	}

	biased_sum_tbX := new(big.Int).Set(share.S.Gx)
	biased_sum_tbY := new(big.Int).Set(share.S.Gy)

	for i := 0; i < len(com_vec2); i++ {
		biased_sum_tbX, biased_sum_tbY = share.PointAdd(biased_sum_tbX, biased_sum_tbY,
			decom_vec2[i].ti.X, decom_vec2[i].ti.Y)
		biased_sum_tbX, biased_sum_tbY = share.PointAdd(biased_sum_tbX, biased_sum_tbY,
			decom_vec1[i].bi.X, decom_vec1[i].bi.Y)
	}
	for i := 0; i < len(com_vec2); i++ {
		biased_sum_tbX, biased_sum_tbY = share.PointSub(
			biased_sum_tbX, biased_sum_tbY,
			decom_vec2[i].ui.X, decom_vec2[i].ui.Y,
		)
	}
	log.Trace(fmt.Sprintf("(gx,gy)=(%s,%s)", share.S.Gx.Text(16), share.S.Gy.Text(16)))
	log.Trace(fmt.Sprintf("(tbx,tby)=(%s,%s)", biased_sum_tbX.Text(16), biased_sum_tbY.Text(16)))
	if share.S.Gx.Cmp(biased_sum_tbX) == 0 &&
		share.S.Gy.Cmp(biased_sum_tbY) == 0 {
		return l.si.Clone(), nil
	}
	return share.PrivKeyZero, errors.New("invalid key")
}

//const
func (l *LocalSignature) outputSignature(s_vec []share.SPrivKey) (r, s share.SPrivKey, err error) {
	s = l.si.Clone()
	for i := 0; i < len(s_vec); i++ {
		share.ModAdd(s, s_vec[i])
	}
	r = share.BigInt2PrivateKey(l.R.X)
	if !verify(s, r, l.y, l.m) {
		err = errors.New("invilad signature")
	}
	return
}

//const
func verify(s, r share.SPrivKey, y *share.SPubKey, message *big.Int) bool {
	b := share.InvertN(s)
	a := share.BigInt2PrivateKey(message)
	u1 := a.Clone()
	u1 = share.ModMul(u1, b)
	u2 := r.Clone()
	u2 = share.ModMul(u2, b)

	gu1x, gu1y := share.S.ScalarBaseMult(u1.Bytes())
	yu2x, yu2y := share.S.ScalarMult(y.X, y.Y, u2.Bytes())
	gu1x, gu1y = share.PointAdd(gu1x, gu1y, yu2x, yu2y)
	if share.BigInt2PrivateKey(gu1x).D.Cmp(r.D) == 0 {
		//return true
	}
	key, _ := crypto.GenerateKey()
	pubkey := key.PublicKey
	pubkey.X = y.X
	pubkey.Y = y.Y
	addr := crypto.PubkeyToAddress(pubkey)
	buf := new(bytes.Buffer)
	buf.Write(utils.BigIntTo32Bytes(r.D))
	buf.Write(utils.BigIntTo32Bytes(s.D))
	buf.Write([]byte{0})
	bs := buf.Bytes()
	h := common.Hash{}
	h.SetBytes(message.Bytes())
	addr2, err := utils.Ecrecover(h, bs)
	if err != nil {
		return false
	}
	if addr2 == addr {
		return true
	}
	bs[64] = 1
	addr2, err = utils.Ecrecover(h, bs)
	if err != nil {
		return false
	}
	if addr2 == addr {
		return true
	}
	return false

}
