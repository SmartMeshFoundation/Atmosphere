package dcrm

import (
	"fmt"

	"math/big"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/models"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/kataras/go-errors"
	"github.com/sirupsen/logrus"
)

type LockedIn struct {
	srv *NotaryService
	db  *models.DB
	Key common.Hash //此次lockedin唯一的key
}

func (l *LockedIn) doLockinKeyGenerate() *models.PrivateKeyInfo {
	pi := &models.PrivateKeyInfo{
		Key: l.Key,
		Zkps: []*models.Zkp{
			{
				Key: l.Key,
			},
		},
	}
	// 构造椭圆曲线上个各项参数================================================================
	// r(rRndS256)是随机数，用来构造私钥k,N为G点的阶(这里eth写法len(N)=256),说明，如果选择其他的加密算法，可以改变r的来源
	rRndS256 := zkp.RandomFromZn(G.N) //阶n=(200-300合适，S256规定是256,再长计算难度大)
	if rRndS256.Sign() == -1 {
		rRndS256.Add(rRndS256, G.P) //GF(mod p)中的p,有限域中的质数
	}
	logrus.Info("选择随机数r=", rRndS256.Bytes())

	// k是私钥，构造成256位
	k := make([]byte, BitSizeOfPrivateKeyShard/8)
	math.ReadBits(rRndS256, k) //凡是涉及到安全的随机,一律用crypto/rand
	logrus.Info("私钥(片)(32字节)=", k)
	//k*G->(Gx,Gy),kG即公钥
	Gx, Gy := secp256k1.S256().ScalarBaseMult(k)
	logrus.Info("公钥(片)Gx=", Gx)
	logrus.Info("公钥(片)Gy=", Gy)

	// 同态加密===============================================================================
	// N为模数（阶）, r(rRndPaillier)随机数
	rRndPaillier := zkp.RandomFromZnStar(l.srv.NotaryShareArg.PaillierPrivateKey.N)
	// 对随机生成的私钥片数据进行加密
	// 输入参数1：Paillier公钥
	// 输入参数2：椭圆曲线的x，即私钥(片)
	// 输入参数3：Paillier选的随机数
	// 输出：加密私钥(片)产生的加密私钥（属于本节点的）encryptK,K是k,即私钥
	encryptK := zkp.Encrypt(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, rRndS256, rRndPaillier)
	//1128
	//encryptK := PaillierPrivateKey.PublicKey.Encrypt(rRndS256).C
	logrus.Info("同态加密结果(加密私钥(片))=", encryptK)

	//h=hash(M可以广播消息)
	var h []*big.Int //len(h)<=4
	marshalGxGy := G.Marshal(Gx, Gy)
	h = append(h, encryptK)                           //1（保存有加密私钥(片)的）
	h = append(h, new(big.Int).SetBytes(marshalGxGy)) //2（保存有公钥的）

	// 提交commit运算pbc来进行双线性配对
	// 输入参数1：随机数
	// 输入参数2：同态加密引擎pbc生成的“权限系统参数”（授权节点才可拥有，相当于私钥分配权利人的证书，发布一次后不可更改，否则checkcommit无法验证，即非法的commit peer）
	// 输入参数3：h,即用于广播出去其他人验证的秘密（含有本节点负责产生的加密私钥（片）和对应公钥）
	// 输出：pbc引擎（共识算法）发方的陷门承结果（1、commitment(含对应pbc形式公钥)，2、open(含pbc形式(hash)的密文)）
	mc := commitments.MultiLinnearCommit(l.srv.NotaryShareArg.MasterPK, h)

	zkParmsOfLockin := new(zkp.Zkp)
	zkParmsOfLockin.Initialization(l.srv.NotaryShareArg.ZkPublicParams,
		rRndS256,
		G.Gx, G.Gy,
		encryptK,
		rRndPaillier)

	z := pi.Zkps[0] //第一个总是我自己私钥片对应的信息
	z.PublicKeyX = Gx
	z.PublicKeyY = Gy
	z.EncryptedK = encryptK
	z.Zkp = zkParmsOfLockin
	z.Cmt = mc
	z.NotaryName = l.srv.NotaryShareArg.Name
	pi.K = rRndS256
	pi.RndPaillier = rRndPaillier
	return pi
}
func (l *LockedIn) LockinKeyGenerate() (*models.Zkp, error) {
	pi := l.doLockinKeyGenerate()
	//err := models.AddZkp(pi.Zkps[0])
	//if err != nil {
	//	return nil, err
	//}
	err := l.db.NewPrivateKeyInfo(pi)
	return pi.Zkps[0], err
}

//收到了来自其他公证人关于此次秘钥生成的结果
func (l *LockedIn) AddNewZkpFromOtherNotary(z *models.Zkp) (finish bool, err error) {
	if !isZkpAllSet(z) {
		err = errors.New("zkp all field must not be nil ")
		return
	}
	pi, err := l.db.LoadPrivatedKeyInfo(l.Key)
	if err != nil {
		return
	}
	if pi.EncryptedPrivateKey != nil {
		err = errors.New(fmt.Sprintf("zkp already verified. key=%s", z.Key.String()))
		return
	}
	//验证公证人是否在重复提交
	for _, z2 := range pi.Zkps {
		if z2.NotaryName == z.NotaryName {
			err = errors.New(fmt.Sprintf("zkp already exist, notary=%s,key=%s", z.NotaryName, z.Key.String()))
			return
		}
	}
	err = l.db.AddZkp(z)
	if err != nil {
		return
	}
	pi.Zkps = append(pi.Zkps, z)
	//所有公证人的证明都集齐了.
	if len(pi.Zkps) == len(l.srv.Notaries) {
		err = l.zkProof(pi.Zkps)
		if err != nil {
			return false, err
		}
		//如果没有零知识证明,没法保证公证人给出的加密私钥片之和与公钥之间的关系.
		pi.PublicKeyX, pi.PublicKeyY = calculatePubKey(pi.Zkps)
		pi.EncryptedPrivateKey = calculateEncPrivateKey(&l.srv.NotaryShareArg.PaillierPrivateKey.PublicKey, pi.Zkps)
		err = l.db.UpdatePrivatedKeyInfoAfterAllZkp(pi)
		return true, err
	}
	return false, nil
}
func (l *LockedIn) zkProof(zkps []*models.Zkp) (err error) {
	//校验commitments
	for _, z := range zkps {
		if !commitments.CheckCommitment(z.Cmt.Commitment, z.Cmt.Open, l.srv.NotaryShareArg.MasterPK) {
			return fmt.Errorf("check commit error for %s", z.NotaryName)
		}
	}
	for _, z := range zkps {
		rr := z.Cmt.Open.GetSecrets()[1]
		rrlen := (rr.BitLen() + 7) / 8
		rrs := make([]byte, rrlen)
		math.ReadBits(rr, rrs)
		rx, ry := G.Unmarshal(rrs)
		//参数：
		//1、共识参数
		//2、秘密
		if !z.Zkp.Verify(l.srv.NotaryShareArg.ZkPublicParams,
			rx, ry,
			z.Cmt.Open.GetSecrets()[0]) {
			return fmt.Errorf("zkp verify err for %s", z.NotaryName)
		}
	}
	return nil
}
func isZkpAllSet(z *models.Zkp) bool {
	z2 := z.Zkp
	return z.EncryptedK != nil &&
		z.PublicKeyX != nil &&
		z.PublicKeyY != nil &&
		z2 != nil &&
		(z2.E != nil &&
			z2.Z != nil &&
			z2.S1 != nil &&
			z2.S2 != nil &&
			z2.S3 != nil &&
			z2.U1 != nil &&
			z2.U2 != nil &&
			z2.U3 != nil)
}

//生成公钥，随便打乱元素顺序 同态加
func calculatePubKey(peers []*models.Zkp) (*big.Int, *big.Int) {
	x := peers[0].PublicKeyX
	y := peers[0].PublicKeyY
	for i := 1; i < len(peers); i++ {
		x, y = G.Add(x, y, peers[i].PublicKeyX, peers[i].PublicKeyY)
	}
	return x, y
}

//生成私钥 同态加
func calculateEncPrivateKey(pubKey *zkp.PublicKey, peers []*models.Zkp) *big.Int {
	encX := peers[0].EncryptedK
	for i := 1; i < len(peers); i++ {
		encX = zkp.CipherAdd(pubKey, encX, peers[i].EncryptedK)
	}
	return encX
}

func calcAddr(X, Y *big.Int) common.Address {
	key, _ := crypto.GenerateKey()
	pub := key.PublicKey
	pub.X = X
	pub.Y = Y
	addr := crypto.PubkeyToAddress(pub)
	return addr
}
