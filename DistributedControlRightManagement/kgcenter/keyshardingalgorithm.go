package kgcenter

import (
	"container/list"

	"github.com/ethereum/go-ethereum/common/math"
	"crypto/rand"
	"github.com/sirupsen/logrus"
	"math/big"
	mathrand "math/rand"
	"time"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"encoding/hex"
	"github.com/tendermint/go-crypto/tmhash"
	"github.com/SmartMeshFoundation/DCRM/util"
	"github.com/DistributedControlRightManagement/configs"
)

//lock -in
func LockIn() {
	InitDcrmList()
	//1、生成我本次负责的片（公私钥）
	logrus.Info("*********************************************LOCK IN************************************************")
	for i := 0; i < configs.ThresholdNum; i++ {
		logrus.Info("[LOCK-IN]（step 1）生成key,Peer ", i+1)
		LockinKeyGenerate()
		logrus.Info("************************************************************************************************")
	}
	//2、零知识证明
	ZkProof(dcrmList)

}

func LockOut()  {
	logrus.Info("*********************************************LOCK OUT************************************************")
	mappingReq:="013d55da472b04a674955d947932d3aee25beba25a7cea60b4d7472d98a76d09"
	signature := Sign(dcrmList,EncX,mappingReq)
	logrus.Info("[LOCK-OUT]ECDSA(R,S,V) R=", signature.r)
	logrus.Info("[LOCK-OUT]ECDSA(R,S,V) S=", signature.s)
	logrus.Info("[LOCK-OUT]ECDSA(R,S,V) V=", signature.GetRecoveryParam())
	if signature != nil {
		if signature.verify(mappingReq,PkX,PkY){
			logrus.Info("[LOCK-OUT]Signature verified passed")
		}else {
			logrus.Info("[LOCK-OUT]Signature verified not passed")
		}
	}else {
		logrus.Error("[LOCK-OUT]signature is null")
	}
}


var dcrmList *list.List
var EncX *big.Int
var PkX,PkY *big.Int
func InitDcrmList() {
	dcrmList=list.New()
	return
}

//共识参数-私钥片的长度
var BitSizeOfPrivateKeyShard=256

/*T=(p,a,b,G,n,h)。
p a b 曲线
p 越大计算越慢 200位左右
p!=n*h
pt!=1(mod n) 1<=t<=20
4a3+27b2!=0(mod p)
G 基点
n G的阶 素数
h 曲线上所有点的个数/n  h<=4
*/
var peerInfo map[int]*ProverInfo

func LockinKeyGenerate() {
	// 构造椭圆曲线上个各项参数================================================================
	// r(rRndS256)是随机数，用来构造私钥k,N为G点的阶(这里eth写法len(N)=256),说明，如果选择其他的加密算法，可以改变r的来源
	rRndS256 := RandomFromZn(configs.G.N) //阶n=(200-300合适，S256规定是256,再长计算难度大)
	if rRndS256.Sign() == -1 {
		rRndS256.Add(rRndS256, configs.G.P) //GF(mod p)中的p,有限域中的质数
	}
	logrus.Info("选择随机数r=", rRndS256.Bytes())
	// k是私钥，构造成256位
	k := make([]byte, BitSizeOfPrivateKeyShard/8)

	math.ReadBits(rRndS256, k)
	logrus.Info("私钥(片)(32字节)=", k)
	//k*G->(Gx,Gy),kG即公钥
	Gx, Gy := configs.G.ScalarBaseMult(k)
	logrus.Info("公钥(片)Gx=", Gx)
	logrus.Info("公钥(片)Gy=", Gy)

	// 同态加密===============================================================================
	// N为模数（阶）, r(rRndPaillier)随机数
	rRndPaillier := RandomFromZnStar(PaillierPrivateKey.N)
	// 对随机生成的私钥片数据进行加密
	// 输入参数1：Paillier公钥
	// 输入参数2：椭圆曲线的x，即私钥(片)
	// 输入参数3：Paillier选的随机数
	// 输出：加密私钥(片)产生的加密私钥（属于本节点的）encryptK,K是k,即私钥
	encryptK :=encrypt(&PaillierPrivateKey.PublicKey, rRndS256, rRndPaillier)
	logrus.Info("同态加密结果(加密私钥(片))=",encryptK)

	//h=hash(M可以广播消息)
	h := []*big.Int{} //len(h)<=4
	marshalGxGy := configs.G.Marshal(Gx, Gy)
	h = append(h, encryptK)                           //1（保存有加密私钥(片)的）
	h = append(h, new(big.Int).SetBytes(marshalGxGy)) //2（保存有公钥的）

	// 提交commit运算pbc来进行双线性配对
	// 输入参数1：随机数
	// 输入参数2：同态加密引擎pbc生成的“权限系统参数”（授权节点才可拥有，相当于私钥分配权利人的证书，发布一次后不可更改，否则checkcommit无法验证，即非法的commit peer）
	// 输入参数3：h,即用于广播出去其他人验证的秘密（含有本节点负责产生的加密私钥（片）和对应公钥）
	// 输出：pbc引擎（共识算法）发方的陷门承结果（1、commitment(含对应pbc形式公钥)，2、open(含pbc形式(hash)的密文)）
	mc := MultiLinnearCommit(SecureRnd, masterPK, h)
	mcOpen := mc.CmtOpen()
	mcCmt := mc.CmtCommitment()
	//保存本节点本此任务的私钥、加密信息来源等========================================================
	tmpPeer := new(ProverInfo)
	tmpPeer.setxShare(rRndS256)
	tmpPeer.setyShare_x(Gx)
	tmpPeer.setyShare_y(Gy)
	tmpPeer.setxShareRnd(rRndPaillier)
	tmpPeer.setEncXShare(encryptK)
	tmpPeer.setMpkEncXiYi(mc)
	tmpPeer.setOpenEncXiYi(mcOpen)//广播
	tmpPeer.setCmtEncXiYi(mcCmt)//广播
	dcrmList.PushBack(tmpPeer)
}

var SecureRnd = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
var PaillierPrivateKey, _= GenerateKey(rand.Reader, 1024)
var zkPublicParams=GenerateParams(configs.G,256, 512, SecureRnd, &PaillierPrivateKey.PublicKey)
var masterPK=GenerateMasterPK()

//1
func ZkProof(peers *list.List) {
	var a1=0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*ProverInfo)
		zkParmsOfLockin := new(Zkp)
		zkParmsOfLockin.Initialization(zkPublicParams,
			peer.getxShare(),
			SecureRnd,
			configs.G.Gx,
			configs.G.Gy,
			peer.getEncXShare(),
			peer.getxShareRnd())
		peer.setZkpKG(zkParmsOfLockin)//广播给其他机器
		logrus.Info("[LOCK-IN]（step 1）零知识证明,设置证明人计算参数,peer ",a1+1)
		a1++
	}

	//check commitment
	//秘密分发MultiTrapdoorCommitment(commitment,open),masterPK是对有权限签发的节点授权的
	var a2=0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer:=e.Value.(*ProverInfo)
		if (Checkcommitment(peer.getCmtEncXiYi(), peer.getOpenEncXiYi(), masterPK) == false) {
			logrus.Fatal("[LOCK-IN]（step 2）Commit验证时发生错误")
		}else {
			logrus.Info("[LOCK-IN]（step 2）Commit验证通过,peer ",a2+1)
		}
		a2++
	}

	//check zero knowledge
	var a3=0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*ProverInfo)
		rr:=peer.getOpenEncXiYi().GetSecrets()[1]
		rrlen:=(rr.BitLen()+7)/8
		rrs:=make([]byte,rrlen)
		math.ReadBits(rr,rrs)
		rx,ry:= configs.G.Unmarshal(rrs)
		//参数：
		//1、共识参数
		//2、秘密
		if peer.getZkpKG().Verify(zkPublicParams,
			rx,ry,
			peer.getOpenEncXiYi().GetSecrets()[0]) {
			logrus.Info("[LOCK-IN]（step 3）零知识证明通过校验,peer ",a3+1)
		}else {
			logrus.Fatal("[LOCK-IN]（step 3）零知识证明校验时未通过,peer ",a3+1)
		}
		a3++
	}

	//保存生成地址
	var a4=0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*ProverInfo)
		encPrivate:=calculateEncPrivateKey(peers)
		peer.setEncX(encPrivate);
		pkx,pky := calculatePubKey(peers)
		peer.setPk_x(pkx);//Pk =pulic key
		peer.setPk_y(pky);
		logrus.Info("************************************************************")
		logrus.Info("同态(加法)私钥片的结果,peer ",a4,",私钥长度：",encPrivate.BitLen()/8,"字节,私钥：",encPrivate)
		logrus.Info("公钥元素x(32):",pkx.Bytes())
		logrus.Info("公钥元素y(32):",pky.Bytes())
		addrBytes := new([64]byte)
		copy(addrBytes[0:32], pkx.Bytes())
		copy(addrBytes[:32], pky.Bytes())
		logrus.Info("分布式", configs.ThresholdNum,"个可信的授权节点分配出的地址:=",hex.EncodeToString(tmhash.Sum(addrBytes[:])),",长度:", len(tmhash.Sum(addrBytes[:])))
		a4++

		EncX=encPrivate
		PkX=pkx
		PkY=pky
	}
	//钱包要比对一下，所有证明人的合成的公私钥要一致
}

//生成公钥
func calculatePubKey(peers *list.List) (*big.Int,*big.Int) {
	yShare_x0 := ((peers.Front().Value).(*ProverInfo)).getyShare_x()
	yShare_y0 := ((peers.Front().Value).(*ProverInfo)).getyShare_y()

	e := peers.Front()
	for e = e.Next(); e != nil; e = e.Next() {
		yShare_xi := ((e.Value).(*ProverInfo)).getyShare_x()
		yShare_yi := ((e.Value).(*ProverInfo)).getyShare_y()

		yShare_x0,yShare_y0 = secp256k1.S256().Add(yShare_x0,yShare_y0,yShare_xi,yShare_yi)
	}
	return yShare_x0,yShare_y0
}

//生成私钥
func calculateEncPrivateKey(peers *list.List) *big.Int {
	encX := ((peers.Front().Value).(*ProverInfo)).getEncXShare()
	e := peers.Front()
	for e = e.Next(); e != nil; e = e.Next() {
		encXi := ((e.Value).(*ProverInfo)).getEncXShare()
		encX = cipherAdd((&PaillierPrivateKey.PublicKey), encX, encXi); //+c1 c2
	}
	return encX
}
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//lock-out
//1
func LockoutCalcCommitment(peers *list.List,encX *big.Int) {
	var rhoI, rhoIRnd, uI, vI *big.Int
	var mpkUiVi *MultiTrapdoorCommitment
	var openUiVi *Open
	var cmtUiVi *Commitment
	a := 0
	for e := peers.Front(); e != nil; e = e.Next() {
		rhoI = util.RandomFromZn(secp256k1.S256().N, SecureRnd)
		rhoIRnd = util.RandomFromZnStar((&PaillierPrivateKey.PublicKey).N, SecureRnd)
		uI = encrypt((&PaillierPrivateKey.PublicKey), rhoI, rhoIRnd)
		vI = cipherMultiply((&PaillierPrivateKey.PublicKey), encX, rhoI)

		var nums= []*big.Int{uI, vI}
		mpkUiVi = MultiLinnearCommit(SecureRnd, masterPK, nums)
		openUiVi = mpkUiVi.CmtOpen()
		cmtUiVi = mpkUiVi.CmtCommitment()

		peer := e.Value.(*ProverInfo)
		peer.setRhoI(rhoI)
		peer.setRhoIRnd(rhoIRnd)
		peer.setuI(uI)
		peer.setvI(vI)
		peer.setMpkUiVi(mpkUiVi)
		peer.setOpenUiVi(openUiVi)
		peer.setCmtUiVi(cmtUiVi)

		logrus.Info("[LOCK-OUT]（step 1）计算Commit,peer ", a+1)
		a = a + 1
	}
}

//2
func LockoutCalcZeroKnowledgeProverI1(peers *list.List,encX *big.Int)  {
	var a1=0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*ProverInfo)
		zkParmsOfLockouti1 := new(Zkpi1)
		zkParmsOfLockouti1.Initialization(
			zkPublicParams,
			peer.getRhoI(),
			SecureRnd,
			peer.getRhoIRnd(),
			peer.getvI(),
			encX,
			peer.getuI())
		peer.setZkp1(zkParmsOfLockouti1)
		logrus.Info("[LOCK-OUT]（step 2）零知识证明i1,设置证明人计算参数,peer ",a1+1)
		a1++
	}
}

//3
func LockoutCheckCommitment(peers *list.List)  bool{
	var a=0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer:=e.Value.(*ProverInfo)
		if (Checkcommitment(peer.getCmtUiVi(), peer.getOpenUiVi(), masterPK) == false) {
			logrus.Fatal("[LOCK-OUT]（step 3）Commit验证时发生错误")
			return false
		}else {
			logrus.Info("[LOCK-OUT]（step 3）Commit验证通过,peer ",a+1)
		}
		a++
	}
	return true
}

//4
func LockoutVerifyZeroKnowledgeI1(peers *list.List,encX *big.Int) bool {
	var a= 0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*ProverInfo)
		if peer.getZkp1().verify(zkPublicParams, configs.G,
			peer.getOpenUiVi().GetSecrets()[1],
			encX,
			peer.getOpenUiVi().GetSecrets()[0],
		) {
			logrus.Info("[LOCK-OUT]（step 4）零知识证明通过校验i1,peer ", a+1)

		} else {
			logrus.Fatal("[LOCK-OUT]（step 4）零知识证明校验i1未通过,peer ", a+1)
			return false
		}
		a++
	}
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////


//5
func LockoutCalcCommitmentOfSign(peers *list.List)  {
	u := calculateU(peers)
	v := calculateV(peers)
	if v.Cmp(big.NewInt(0))==0{
		logrus.Warn("V is 0")
	}
	a:= 0
	for e := peers.Front(); e != nil; e = e.Next() {
		kI := util.RandomFromZn(secp256k1.S256().N, SecureRnd)
		if kI.Sign() == -1 {
			kI.Add(kI,secp256k1.S256().P)
		}
		rI := make([]byte, 32)
		math.ReadBits(kI, rI[:])
		rIx,rIy := KMulG(rI[:])
		cI := util.RandomFromZn(secp256k1.S256().N, SecureRnd)
		cIRnd := util.RandomFromZnStar((&PaillierPrivateKey.PublicKey).N,SecureRnd)
		mask := encrypt((&PaillierPrivateKey.PublicKey),new(big.Int).Mul(secp256k1.S256().N, cI),cIRnd)
		wI := cipherAdd((&PaillierPrivateKey.PublicKey),cipherMultiply((&PaillierPrivateKey.PublicKey),u, kI), mask)
		rIs := secp256k1.S256().Marshal(rIx,rIy)

		var nums = []*big.Int{new(big.Int).SetBytes(rIs[:]),wI}
		mpkRiWi := MultiLinnearCommit(SecureRnd,masterPK,nums)

		openRiWi := mpkRiWi.CmtOpen()
		cmtRiWi := mpkRiWi.CmtCommitment()

		peer := e.Value.(*ProverInfo)
		peer.setkI(kI)
		peer.setcI(cI)
		peer.setcIRnd(cIRnd)
		peer.setrI_x(rIx)
		peer.setrI_y(rIy)
		peer.setMask(mask)
		peer.setwI(wI)
		peer.setMpkRiWi(mpkRiWi)
		peer.setOpenRiWi(openRiWi)
		peer.setCmtRiWi(cmtRiWi)
		logrus.Info("[LOCK-OUT]（step 5）计算签名的commit,peer ",a+1)
		a++
	}
}

//6
func LockoutCalcZeroKnowledgeProverI2OfSign(peers *list.List)  {
	a:=0
	u := calculateU(peers)
	for e := peers.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*ProverInfo)
		zkParmsOfLockouti2 := new(Zkpi2)
		zkParmsOfLockouti2.Initialization(
			zkPublicParams,
			peer.getkI(),
			peer.getcI(),
			SecureRnd,
			configs.G.Gx,
			configs.G.Gy,
			peer.getwI(),
			u,
			peer.getcIRnd())
		peer.setZkp_i2(zkParmsOfLockouti2)
		logrus.Info("[LOCK-OUT]（step 6）零知识证明i2,设置签名的证明人计算参数,peer ",a+1)
		a++
	}
}

//7
func LockoutCheckCommitmentOfsign(peers *list.List) bool{
	a := 0
	for e := peers.Front(); e != nil; e = e.Next() {
		peer:=e.Value.(*ProverInfo)
		if Checkcommitment(peer.getCmtRiWi(), peer.getOpenRiWi(),masterPK) == false {
			logrus.Fatal("[LOCK-OUT]（step 7）Commit验证时发生错误")
			return false
		}
		logrus.Info("[LOCK-OUT]（step 7）校验commit,peer ",a+1)
		a = a+1
	}
	return true
}

//8
func LockoutVerifyZeroKnowledgeI2OfSign(peers *list.List,u *big.Int) bool {
	a := 0
	for e := peers.Front(); e != nil; e = e.Next() {

		rr := ((e.Value).(*ProverInfo)).getOpenRiWi().GetSecrets()[0]
		rrlen := ((rr.BitLen()+7)/8)
		rrs := make([]byte,rrlen)
		math.ReadBits(rr,rrs[:])
		rx,ry := secp256k1.S256().Unmarshal(rrs[:])
		peer:=e.Value.(*ProverInfo)
		if peer.getZkp_i2().verify(zkPublicParams,secp256k1.S256(),
			rx,ry,u,peer.getOpenRiWi().GetSecrets()[1]) == false {
			logrus.Info("[LOCK-OUT]（step 8）零知识证明校验i2未通过,peer ", a+1)
			return false
		}else {
			logrus.Info("[LOCK-OUT]（step 8）零知识证明通过校验i2,peer ", a+1)
		}
		a++
	}
	return true
}


//9
func LockoutCalcSignature(peers *list.List,u *big.Int,v *big.Int,message string) *ECDSASignature {
	signature := new(ECDSASignature)
	signature.New()

	w := calculateW(peers)
	rx,ry := calculateR(peers)

	r := new(big.Int).Mod(rx,secp256k1.S256().N)
	mu := decrypt(PaillierPrivateKey,w)
	mu.Mod(mu,secp256k1.S256().N)
	muInverse := new(big.Int).ModInverse(mu, configs.G.N)
	msgDigest,_ := new(big.Int).SetString(message,16)
	mMultiU := cipherMultiply((&PaillierPrivateKey.PublicKey),u, msgDigest)
	rMultiV := cipherMultiply((&PaillierPrivateKey.PublicKey),v, r)
	sEnc := cipherMultiply((&PaillierPrivateKey.PublicKey),cipherAdd((&PaillierPrivateKey.PublicKey),mMultiU, rMultiV), muInverse)

	s := decrypt(PaillierPrivateKey,sEnc)
	s.Mod(s,secp256k1.S256().N)

	signature.setR(r)
	signature.setS(s)

	two,_ := new(big.Int).SetString("2",10)
	ryy := new(big.Int).Mod(ry,two)
	zero,_ := new(big.Int).SetString("0",10)
	cmp := ryy.Cmp(zero)
	recoveryParam := 1
	if cmp == 0 {
		recoveryParam = 0
	}

	tt := new(big.Int).Rsh(secp256k1.S256().N,1)
	comp := s.Cmp(tt)
	if comp > 0 {
		recoveryParam = 1
		s = new(big.Int).Sub(secp256k1.S256().N,s)
		signature.setS(s);
	}
	//test
	signature.setRecoveryParam(int32(recoveryParam))
	return signature

}

func Sign(peers *list.List,encX *big.Int,message string) *ECDSASignature {
	//所有负责节点来验证公钥和加密的私钥
	LockoutCalcCommitment(peers,encX)
	LockoutCalcZeroKnowledgeProverI1(peers,encX)

	if LockoutCheckCommitment(peers)== false {
		return nil
	}
	if LockoutVerifyZeroKnowledgeI1(peers,encX)== false {
		return nil
	}

	u := calculateU(peers)
	v := calculateV(peers)
	LockoutCalcCommitmentOfSign(peers)
	LockoutCalcZeroKnowledgeProverI2OfSign(peers)

	if LockoutCheckCommitmentOfsign(peers)==false{
		return nil
	}
	if LockoutVerifyZeroKnowledgeI2OfSign(peers,u)==false{
		return nil
	}
	//生成签名
	sinature:=LockoutCalcSignature(peers,u,v,message)
	return sinature
}

func calculateW(peers *list.List) *big.Int {
	w := ((peers.Front().Value).(*ProverInfo)).getOpenRiWi().GetSecrets()[1]
	e := peers.Front()
	for e = e.Next(); e != nil; e = e.Next() {
		wi := ((e.Value).(*ProverInfo)).getOpenRiWi().GetSecrets()[1]
		w = cipherAdd((&PaillierPrivateKey.PublicKey),w,wi);
	}
	//fmt.Println("Calculate the Encrypted Inner-Data w: ",w)
	return w
}

func calculateR(peers *list.List) (*big.Int,*big.Int) {

	rr := ((peers.Front().Value).(*ProverInfo)).getOpenRiWi().GetSecrets()[0]
	rrlen := ((rr.BitLen()+7)/8)
	rrs := make([]byte,rrlen)
	math.ReadBits(rr,rrs[:])
	rx,ry := secp256k1.S256().Unmarshal(rrs[:])

	e := peers.Front()
	for e = e.Next(); e != nil; e = e.Next() {

		rri := ((e.Value).(*ProverInfo)).getOpenRiWi().GetSecrets()[0]
		rrilen := ((rri.BitLen()+7)/8)
		rris := make([]byte,rrilen)
		math.ReadBits(rri,rris[:])
		rrix,rriy := secp256k1.S256().Unmarshal(rris[:])

		rx,ry = secp256k1.S256().Add(rx,ry,rrix,rriy)
	}
	return rx,ry
}

func calculateU(peers *list.List) *big.Int {
	u := ((peers.Front().Value).(*ProverInfo)).getOpenUiVi().GetSecrets()[0]
	e := peers.Front()
	for e = e.Next(); e != nil; e = e.Next() {
		ui := ((e.Value).(*ProverInfo)).getOpenUiVi().GetSecrets()[0]
		u = cipherAdd((&PaillierPrivateKey.PublicKey),u,ui)
	}
	return u
}

func calculateV(userList *list.List) *big.Int {
	v := ((userList.Front().Value).(*ProverInfo)).getOpenUiVi().GetSecrets()[1]
	e := userList.Front()
	for e = e.Next(); e != nil; e = e.Next() {
		vi := ((e.Value).(*ProverInfo)).getOpenUiVi().GetSecrets()[1]
		v = cipherAdd((&PaillierPrivateKey.PublicKey),v,vi)
	}
	return v
}

