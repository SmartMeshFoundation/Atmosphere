package models

import (
	"math/big"

	"errors"

	"bytes"
	"encoding/gob"
	"fmt"

	"encoding/json"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
)

var errKeyLength = errors.New("key length error")

/*
lockedin 过程中互相之间协商的结果
*/
type PrivateKeyInfo struct {
	Key                 common.Hash
	PublicKeyX          *big.Int // 此次协商生成的私钥对应的公钥 X,Y *big.Int
	PublicKeyY          *big.Int
	EncryptedPrivateKey *big.Int //加密后的总私钥 *big.Int
	K                   *big.Int //32字节的私钥片 ,lockout时候这个是关键
	RndPaillier         *big.Int //同态加密参数 rRndPaillier
	Zkps                []*Zkp
}
type PrivateKeyInfoModel struct {
	Key                 []byte `gorm:"primary_key"` //a random hash
	PublicKeyX          string // 此次协商生成的私钥对应的公钥 X,Y *big.Int
	PublicKeyY          string
	EncryptedPrivateKey string      //加密后的总私钥 *big.Int
	K                   string      //32字节的私钥片 ,lockout时候这个是关键
	RndPaillier         string      //同态加密参数 rRndPaillier
	Zkps                []*ZkpModel `gorm:"ForeignKey:Key"`
}

type Zkp struct {
	Key        common.Hash
	EncryptedK *big.Int //加密后的私钥片  //*big.Int
	PublicKeyX *big.Int //私钥片对应的公钥
	PublicKeyY *big.Int //私钥片对应的公钥
	Zkp        *zkp.Zkp //用于生成总公钥的,总加密私钥的信息
	Cmt        *commitments.MultiTrapdoorCommitment
	NotaryName string
}

type ZkpModel struct {
	ID         int    `json:"-"`     //ignore
	Key        []byte `gorm:"index"` //PrivateKeyInfo 协商时用的key
	EncryptedK string //加密后的私钥片  //*big.Int
	PublicKeyX string //私钥片对应的公钥
	PublicKeyY string //私钥片对应的公钥
	Zkp        []byte `gorm:"size:512"` //用于生成总公钥的零知识证明,需要存储10个大整数.
	Cmt        []byte `gorm:"size:1024"`
	NotaryName string
}

func strToBigInt(s string) *big.Int {
	if len(s) > 0 {
		i := new(big.Int)
		i.SetString(s, 0)
		return i
	}
	return nil
}
func bigIntToStr(i *big.Int) string {
	if i != nil {
		return i.String()
	}
	return ""
}
func fromPrivateKeyInfoModel(p *PrivateKeyInfoModel) *PrivateKeyInfo {
	p2 := &PrivateKeyInfo{
		PublicKeyX:          strToBigInt(p.PublicKeyX),
		PublicKeyY:          strToBigInt(p.PublicKeyY),
		EncryptedPrivateKey: strToBigInt(p.EncryptedPrivateKey),
		K:                   strToBigInt(p.K),
		RndPaillier:         strToBigInt(p.RndPaillier),
	}
	p2.Key.SetBytes(p.Key)
	for _, z := range p.Zkps {
		z2 := fromZkpModel(z)
		p2.Zkps = append(p2.Zkps, z2)
	}
	return p2
}
func toPrivateKeyInfoModel(p *PrivateKeyInfo) *PrivateKeyInfoModel {
	p2 := &PrivateKeyInfoModel{
		Key:                 p.Key[:],
		PublicKeyX:          bigIntToStr(p.PublicKeyX),
		PublicKeyY:          bigIntToStr(p.PublicKeyY),
		EncryptedPrivateKey: bigIntToStr(p.EncryptedPrivateKey),
		K:                   bigIntToStr(p.K),
		RndPaillier:         bigIntToStr(p.RndPaillier),
	}
	for _, z := range p.Zkps {
		z2 := toZkpModel(z)
		p2.Zkps = append(p2.Zkps, z2)
	}
	return p2
}
func fromZkpModel(z *ZkpModel) *Zkp {
	z2 := &Zkp{
		EncryptedK: strToBigInt(z.EncryptedK),
		PublicKeyX: strToBigInt(z.PublicKeyX),
		PublicKeyY: strToBigInt(z.PublicKeyY),
		NotaryName: z.NotaryName,
	}
	z2.Key.SetBytes(z.Key)
	if len(z.Zkp) > 10 { //gorm在处理空的zkp的时候有问题. 长度是1,内容为0
		d := gob.NewDecoder(bytes.NewBuffer(z.Zkp))
		z2.Zkp = &zkp.Zkp{}
		err := d.Decode(z2.Zkp)
		if err != nil {
			panic(fmt.Sprintf("decode z  err %s,  %s", err, utils.StringInterface(z, 3)))
		}
	}
	if len(z.Cmt) > 10 {
		z2.Cmt = &commitments.MultiTrapdoorCommitment{}
		err := json.Unmarshal(z.Cmt, z2.Cmt)
		if err != nil {
			panic(fmt.Sprintf("decode MultiTrapdoorMasterPublicKey err %s ", err))
		}
	}
	return z2
}
func toZkpModel(z *Zkp) *ZkpModel {
	z2 := &ZkpModel{
		Key:        z.Key[:],
		EncryptedK: bigIntToStr(z.EncryptedK),
		PublicKeyX: bigIntToStr(z.PublicKeyX),
		PublicKeyY: bigIntToStr(z.PublicKeyY),
		NotaryName: z.NotaryName,
	}
	if z.Zkp != nil {
		b := new(bytes.Buffer)
		e := gob.NewEncoder(b)
		err := e.Encode(z.Zkp)
		if err != nil {
			panic(fmt.Sprintf("gob encode error %s, %s", err, utils.StringInterface(z, 3)))
		}
		z2.Zkp = b.Bytes()
	}
	if z.Cmt != nil {
		var err error
		z2.Cmt, err = json.Marshal(z.Cmt)
		if err != nil {
			panic(fmt.Sprintf("marshal cmt err %s", err))
		}
	}
	return z2
}
func (db *DB) NewPrivateKeyInfo(p *PrivateKeyInfo) error {
	return db.Create(toPrivateKeyInfoModel(p)).Error
}

func (db *DB) LoadPrivatedKeyInfo(key common.Hash) (*PrivateKeyInfo, error) {
	var pi PrivateKeyInfoModel
	err := db.Where(&PrivateKeyInfoModel{
		Key: key[:],
	}).Preload("Zkps").Find(&pi).Error
	if err != nil {
		return nil, err
	}
	return fromPrivateKeyInfoModel(&pi), nil
}

//添加一个新的zp到数据库中,zkp是不用更新的,一旦写入不会更改
func (db *DB) AddZkp(z *Zkp) error {
	return db.Create(toZkpModel(z)).Error
}

func (db *DB) UpdatePrivatedKeyInfoAfterAllZkp(pi *PrivateKeyInfo) error {
	return db.Model(&PrivateKeyInfoModel{
		Key: pi.Key[:],
	}).Update(&PrivateKeyInfoModel{
		PublicKeyX:          bigIntToStr(pi.PublicKeyX),
		PublicKeyY:          bigIntToStr(pi.PublicKeyY),
		EncryptedPrivateKey: bigIntToStr(pi.EncryptedPrivateKey),
	}).Error
}
