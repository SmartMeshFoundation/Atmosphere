package dcrm

import (
	"crypto/ecdsa"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

//ListenIP my listen ip
var ListenIP = "127.0.0.1"

//ListenPort my listen port
var ListenPort = "18000"

//NotaryShareArg share info between all notatories.
type NotaryShareArg struct {
	/*
		PaillierPrivateKey 同态加密时候需要在所有公证人之间分享的私钥,在公证人环境初始化的时候创建
		并且公证人应该保护好这个私钥文件,不要泄露任何信息给其他人
	*/
	PaillierPrivateKey *zkp.PrivateKey
	//ZkPublicParams 节点之间进行零知识证明所需的公共参数,在公证人环境初始化的时候创建,以后保持不变
	//并且公证人应该保护好这个参数信息,不要泄露任何信息给其他人
	ZkPublicParams *zkp.PublicParameters
	/*
		MasterPK 是校验commitments的公共public key,在公证人环境初始化的时候创建,以后保持不变
		公证人应该保护好此信息,不要泄露给任何给其他人
	*/
	MasterPK *commitments.MultiTrapdoorMasterPublicKey
	//ThresholdNum 总共有多少个公证人
	ThresholdNum int
	Name         string // name of this notary
}

//NotatoryShareInfo share info between all notatories.
//var NotatoryShareInfo NotaryShareArg

//NotatoryInfo 公证人的基本信息
type NotatoryInfo struct {
	Name string
	Addr string //how to contact with this notary
	Key  *ecdsa.PublicKey
}

//Notaries 除我以外的其他公证人信息
//var Notaries = make(map[string]*NotatoryInfo)

//G 工作的曲线
var G = secp256k1.S256()

//共识参数-私钥片的长度
var BitSizeOfPrivateKeyShard = 256
