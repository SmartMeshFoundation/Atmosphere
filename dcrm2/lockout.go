package dcrm2

import (
	"math/big"

	"fmt"

	"bytes"

	"github.com/SmartMeshFoundation/Atmosphere/DistributedControlRightManagement/configs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/models"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/sirupsen/logrus"
)

type LockedOut struct {
	db                  *models.DB
	srv                 *NotaryService
	Key                 common.Hash //此次lockedin唯一的key
	Message             []byte      //message to sign
	EncryptedPrivateKey *big.Int
	pi                  *models.PrivateKeyInfo
}

//计算lockout stage1 的commitment和zkp
func (l *LockedOut) doCalcCommitmentAndZkpi1() (z *models.Zkpi1, err error) {
	var rhoI, rhoIRnd, uI, vI *big.Int
	var mpkUiVi *commitments.MultiTrapdoorCommitment
	a := 0

	rhoI = zkp.RandomFromZn(G.N)
	rhoIRnd = zkp.RandomFromZnStar(l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey.N)
	uI = zkp.Encrypt(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, rhoI, rhoIRnd)
	vI = zkp.CipherMultiply(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, l.EncryptedPrivateKey, rhoI)
	//1128
	var nums = []*big.Int{uI, vI}
	mpkUiVi = commitments.MultiLinnearCommit(l.srv.NotaryShareArg.MasterPK, nums)
	z = &models.Zkpi1{
		Arg: &models.Zkpi1CommitmentArg{
			RhoI:    rhoI,
			RhoIRnd: rhoIRnd,
			UI:      uI,
			VI:      vI,
		},
		Cmt:        mpkUiVi,
		NotaryName: l.srv.NotaryShareArg.Name,
	}
	zkpLockout1 := new(zkp.Zkpi1)
	zkpLockout1.Initialization(l.srv.NotaryShareArg.ZkPublicParams,
		rhoI,
		rhoIRnd,
		vI,
		l.EncryptedPrivateKey,
		uI)
	z.Zkpi1 = zkpLockout1

	logrus.Info("[LOCK-OUT]（step 1）计算Commit,peer ", a+1)
	return z, nil
}
func (l *LockedOut) NewTx() (tx *models.Tx, err error) {
	z, err := l.doCalcCommitmentAndZkpi1()
	if err != nil {
		return
	}
	tx = &models.Tx{
		Key:     l.Key,
		Message: l.Message,
		Zkpi1s:  []*models.Zkpi1{z},
	}
	err = l.db.NewTx(tx)
	return
}
func (l *LockedOut) newTx2() (tx *models.Tx, err error) {
	z, err := l.doCalcCommitmentAndZkpi1()
	if err != nil {
		return
	}
	tx, err = l.db.LoadTx(l.Key)
	if err == nil {
		tx.Zkpi1s = append(tx.Zkpi1s, z)
		err = l.db.UpdateTxZkpi1(tx)
		return
	}
	tx = &models.Tx{
		Key:     l.Key,
		Message: l.Message,
		Zkpi1s:  []*models.Zkpi1{z},
	}
	err = l.db.NewTx(tx)
	return
}

func (l *LockedOut) AddNewZkpi1FromOtherNotary(z *models.Zkpi1) (finish bool, err error) {
	NotaryNum := len(l.srv.Notaries)
	tx, err := l.db.LoadTx(l.Key)
	if err != nil {
		return
	}
	for _, z2 := range tx.Zkpi1s {
		if z2.NotaryName == z.NotaryName {
			err = fmt.Errorf("zkpi1 already exist for %s", z.NotaryName)
			return
		}
	}
	if !commitments.CheckCommitment(z.Cmt.Commitment, z.Cmt.Open, l.srv.NotaryShareArg.MasterPK) {
		err = fmt.Errorf("commitment for stage1 not pass, notary=%s,key=%s", z.NotaryName, l.Key.String())
		return
	}
	if !z.Zkpi1.Verify(l.srv.NotaryShareArg.ZkPublicParams, G,
		z.Cmt.Open.GetSecrets()[1],
		l.EncryptedPrivateKey,
		z.Cmt.Open.GetSecrets()[0]) {
		err = fmt.Errorf("zkpi1 for stage1 not pass,notary=%s,key=%s", z.NotaryName, l.Key.String())
		return
	}
	tx.Zkpi1s = append(tx.Zkpi1s, z)
	err = l.db.UpdateTxZkpi1(tx)
	if err != nil {
		return
	}
	if len(tx.Zkpi1s) == NotaryNum {
		//zkp stage1 所需数据准备完毕,启动第二步校验
		return true, nil
	}
	return false, nil
}

func (l *LockedOut) calculateU(zs []*models.Zkpi1) *big.Int {
	u := zs[0].Cmt.Open.GetSecrets()[0]
	for i := 1; i < len(zs); i++ {
		ui := zs[i].Cmt.Open.GetSecrets()[0]
		u = zkp.CipherAdd(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, u, ui)
	}
	return u
}

func (l *LockedOut) calculateV(zs []*models.Zkpi1) *big.Int {
	v := zs[0].Cmt.Open.GetSecrets()[1]
	for i := 1; i < len(zs); i++ {
		vi := zs[i].Cmt.Open.GetSecrets()[1]
		v = zkp.CipherAdd(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, v, vi)
	}
	return v
}

func (l *LockedOut) doCalcCommitmentSignAndZkpi2() (z *models.Zkpi2, err error) {
	tx, err := l.db.LoadTx(l.Key)
	if err != nil {
		return
	}
	u := l.calculateU(tx.Zkpi1s)
	v := l.calculateV(tx.Zkpi1s)
	if v.Cmp(big.NewInt(0)) == 0 {
		logrus.Warn("V is 0")
	}
	kI := zkp.RandomFromZn(G.N)
	if kI.Sign() == -1 {
		kI.Add(kI, G.P)
	}
	rI := make([]byte, 32)
	math.ReadBits(kI, rI[:])
	rIx, rIy := zkp.KMulG(rI[:])
	cI := zkp.RandomFromZn(G.N)
	cIRnd := zkp.RandomFromZnStar(l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey.N)
	mask := zkp.Encrypt(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, new(big.Int).Mul(G.N, cI), cIRnd)
	wI := zkp.CipherAdd(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, zkp.CipherMultiply(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, u, kI), mask)
	rIs := secp256k1.S256().Marshal(rIx, rIy)

	var nums = []*big.Int{new(big.Int).SetBytes(rIs[:]), wI}
	mpkRiWi := commitments.MultiLinnearCommit(l.srv.NotaryShareArg.MasterPK, nums)
	z = &models.Zkpi2{
		Arg: &models.Zkpi2CommitmentArg{
			KI:    kI,
			CI:    cI,
			CIRnd: cIRnd,
			RIx:   rIx,
			RIy:   rIy,
			Mask:  mask,
			WI:    wI,
		},
		Cmt:        mpkRiWi,
		NotaryName: l.srv.NotaryShareArg.Name,
	}
	zkParmsOfLockouti2 := new(zkp.Zkpi2)
	zkParmsOfLockouti2.Initialization(
		l.srv.NotaryShareArg.ZkPublicParams,
		kI,
		cI,
		G.Gx,
		G.Gy,
		wI,
		u,
		cIRnd)
	z.Zkpi2 = zkParmsOfLockouti2
	return z, nil
}
func (l *LockedOut) StartLockoutStage2() (tx *models.Tx, err error) {
	tx, err = l.db.LoadTx(l.Key)
	if err != nil {
		return
	}
	z, err := l.doCalcCommitmentSignAndZkpi2()
	if err != nil {
		return
	}
	tx.Zkpi2s = []*models.Zkpi2{z}
	err = l.db.UpdateTxZkpi2(tx)
	return
}

func (l *LockedOut) StartLockoutStage2Test() (tx *models.Tx, err error) {
	tx, err = l.db.LoadTx(l.Key)
	if err != nil {
		return
	}
	z, err := l.doCalcCommitmentSignAndZkpi2()
	if err != nil {
		return
	}
	tx.Zkpi2s = append(tx.Zkpi2s, z)
	err = l.db.UpdateTxZkpi2(tx)
	return
}
func (l *LockedOut) AddZkpi2(z *models.Zkpi2) (tx *models.Tx, finish bool, err error) {
	tx, err = l.db.LoadTx(l.Key)
	if err != nil {
		return
	}
	for _, z2 := range tx.Zkpi2s {
		if z2.NotaryName == z.NotaryName {
			err = fmt.Errorf("zkpi2 already exist for notary=%s,key=%s", z.NotaryName, l.Key.String())
			return
		}
	}
	if !commitments.CheckCommitment(z.Cmt.Commitment, z.Cmt.Open, l.srv.NotaryShareArg.MasterPK) {
		err = fmt.Errorf("zkpi2  commitment not pass,notary=%s", z.NotaryName)
		return
	}
	rr := z.Cmt.Open.GetSecrets()[0]
	rrlen := (rr.BitLen() + 7) / 8
	rrs := make([]byte, rrlen)
	math.ReadBits(rr, rrs[:])
	rx, ry := G.Unmarshal(rrs[:])
	u := l.calculateU(tx.Zkpi1s)
	if !z.Zkpi2.Verify(l.srv.NotaryShareArg.ZkPublicParams, G,
		rx, ry, u, z.Cmt.Open.GetSecrets()[1]) {
		err = fmt.Errorf("zkpi2 not pass,notary=%s", z.NotaryName)
		return
	}
	tx.Zkpi2s = append(tx.Zkpi2s, z)
	err = l.db.UpdateTxZkpi2(tx)
	if err != nil {
		return
	}
	finish = len(tx.Zkpi2s) == len(l.srv.Notaries)
	return
}
func (l *LockedOut) calculateW(zs []*models.Zkpi2) *big.Int {
	w := zs[0].Cmt.Open.GetSecrets()[1]

	for i := 1; i < len(zs); i++ {
		wi := zs[i].Cmt.Open.GetSecrets()[1]
		w = zkp.CipherAdd(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, w, wi)
	}
	return w
}

func (l *LockedOut) calculateR(zs []*models.Zkpi2) (*big.Int, *big.Int) {

	rr := zs[0].Cmt.Open.GetSecrets()[0]
	rrlen := (rr.BitLen() + 7) / 8
	rrs := make([]byte, rrlen)
	math.ReadBits(rr, rrs[:])
	rx, ry := secp256k1.S256().Unmarshal(rrs[:])
	for i := 1; i < len(zs); i++ {
		rri := zs[i].Cmt.Open.GetSecrets()[0]
		rrilen := (rri.BitLen() + 7) / 8
		rris := make([]byte, rrilen)
		math.ReadBits(rri, rris[:])
		rrix, rriy := secp256k1.S256().Unmarshal(rris[:])
		rx, ry = secp256k1.S256().Add(rx, ry, rrix, rriy)
	}
	return rx, ry
}

func isZkpi2AllNonZero(z *zkp.Zkpi2) bool {
	if z.U1 != nil && len(z.U1) == 2 && z.U2 != nil &&
		z.U3 != nil && z.Z1 != nil && z.Z2 != nil &&
		z.S1 != nil && z.S2 != nil && z.T1 != nil &&
		z.T2 != nil && z.T3 != nil && z.E != nil &&
		z.V1 != nil && z.V3 != nil {
		return true
	}
	return false
}
func isZkpi1AllNonZero(z *zkp.Zkpi1) bool {
	return z.Z != nil && z.U1 != nil && z.U2 != nil &&
		z.S1 != nil && z.S2 != nil && z.S3 != nil &&
		z.E != nil && z.V != nil
}
func (l *LockedOut) verifyArg() {
	tx, err := l.db.LoadTx(l.Key)
	if err != nil {
		panic(err)
	}
	for _, z := range tx.Zkpi1s {
		if !isZkpi1AllNonZero(z.Zkpi1) {
			panic(fmt.Sprintf("z=%s", z.NotaryName))
		}
	}
	for _, z := range tx.Zkpi2s {
		if !isZkpi2AllNonZero(z.Zkpi2) {
			panic(fmt.Sprintf("z2=%s", z.NotaryName))
		}
	}
}
func (l *LockedOut) CalcSignature(tx *models.Tx) []byte {
	l.verifyArg()
	tx, err := l.db.LoadTx(l.Key)
	if err != nil {
		panic(err)
	}
	signature := new(zkp.ECDSASignature)
	N := secp256k1.S256().N

	w := l.calculateW(tx.Zkpi2s)
	rx, ry := l.calculateR(tx.Zkpi2s)
	u := l.calculateU(tx.Zkpi1s)
	v := l.calculateV(tx.Zkpi1s)

	r := new(big.Int).Mod(rx, N)
	mu := zkp.Decrypt(l.srv.NotaryShareArg.PaillierPrivateKey, w)
	mu.Mod(mu, secp256k1.S256().N)
	muInverse := new(big.Int).ModInverse(mu, configs.G.N)
	msgDigest := utils.Sha3(l.Message)
	mMultiU := zkp.CipherMultiply(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, u, new(big.Int).SetBytes(msgDigest[:]))
	rMultiV := zkp.CipherMultiply(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, v, r)
	sEnc := zkp.CipherMultiply(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey,
		zkp.CipherAdd(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, mMultiU, rMultiV), muInverse)

	s := zkp.Decrypt(l.srv.NotaryShareArg.PaillierPrivateKey, sEnc)
	s.Mod(s, secp256k1.S256().N)

	signature.SetR(r)
	signature.SetS(s)

	two, _ := new(big.Int).SetString("2", 10)
	ryy := new(big.Int).Mod(ry, two)
	zero, _ := new(big.Int).SetString("0", 10)
	cmp := ryy.Cmp(zero)
	recoveryParam := 1
	if cmp == 0 {
		recoveryParam = 0
	}

	tt := new(big.Int).Rsh(N, 1)
	comp := s.Cmp(tt)
	if comp > 0 {
		recoveryParam = 1
		s = new(big.Int).Sub(N, s)
		signature.SetS(s)
	}
	//test
	signature.SetRecoveryParam(int32(recoveryParam))
	buf := new(bytes.Buffer)
	buf.Write(utils.BigIntTo32Bytes(r))
	buf.Write(utils.BigIntTo32Bytes(s))
	buf.Write([]byte{byte(recoveryParam)})
	log.Info(fmt.Sprintf("calc signature %s x=%s,y=%s", l.srv.NotaryShareArg.Name,
		l.pi.PublicKeyX.String(), l.pi.PublicKeyY.String(),
	))
	if !signature.Verify(l.Message, l.pi.PublicKeyX, l.pi.PublicKeyY) {
		log.Error(fmt.Sprintf("signature verify error for %s", l.srv.NotaryShareArg.Name))
	}
	return buf.Bytes()

}
