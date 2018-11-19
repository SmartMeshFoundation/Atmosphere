package main

import (
	"fmt"
	"crypto/sha256"
	"github.com/Nik-U/pbc"
)

/*
如何用PBC library实现the Boneh-Lynn-Shacham (BLS) signature scheme
基础说明：阶为质数r的三个群G1，G2，GT（定理：阶为质数的群都是循环群,）
定义双线性映射e:G1*G2–>GT，公开G2的一个随机生成元g.
Alice想要对我一个消息签名。她通过如下方法生成公钥和私钥：
私钥：Zr的一个随机元素x
公钥：g^x
为了签名消息，Alice将消息m作为输入，通过哈希算法得到hash值h=hash(m)，对h进行签名sig=h^x，输出sig,发给Bob.
为了验证签名sig,Bob check 双线性映射式子：e(h,g^x) = e(sig, g).是否相等
其中e(h,y)=e(h,g^x)=e(h,g)^x;
若e(sig’,g)=e(sig,g)=e(h^x,g)=e(h,g)^x=e(h,y)，则说明B收到的签名是A的真实签名

*/
//向网络发送的签了名的消息
type messageData struct {
	message   string
	signature []byte
}
var(
	ip1="192.168.124.13"
	ip2="192.168.124.15"
	ip3="192.168.124.2"
	ip4="192.168.124.10"
)
//计算和验证alice和bob模拟通信中的pbc短签名（Boneh-Lynn-Shacham signature）
func main() {
	//权限生成系统参数
	//在实际应用中，只需生成一次并发布它
	params := pbc.GenerateA(160, 512)//
	//实例化一个pair(赋随机参数)
	pairing := params.NewPairing()
	g := pairing.NewG2().Rand()
	//权限管理将参数和g分配给alice和bob，先假定此系统的一个共识，都信任
	sharedParams := params.String()

	sharedG := g.Bytes()
	//模拟出消息的通道，即实际中的网络
	messageChannel := make(chan *messageData)

	//公钥发布通道
	//pbc认为需要一个信任结构来发布公钥
	//公钥只需要传输和验证一次
	keyChannel := make(chan []byte)

	//通道会一直等待指导两个模拟完成
	finished := make(chan bool)

	go alice(sharedParams, sharedG, messageChannel, keyChannel, finished)
	go bob(sharedParams, sharedG, messageChannel, keyChannel, finished)
	//go cindy(sharedParams, sharedG, messageChannel, keyChannel, finishedToCindy)
	//等待通信借宿
	<-finished
	<-finished
}

//alice生成密钥并对消息进行签名
func alice(sharedParams string, sharedG []byte, messageChannel chan *messageData, keyChannel chan []byte, finished chan bool) {
	//alice加载系统公共参数
	//从字符串中加载配对参数并实例化pair
	pairing, _ := pbc.NewPairingFromString(sharedParams)
	gAlice := pairing.NewG2().SetBytes(sharedG)

	//生成 keypair (x, g^x)
	privKey := pairing.NewZr().Rand()
	pubKey := pairing.NewG2().PowZn(gAlice, privKey)

	//向bob发送公钥
	keyChannel <- pubKey.Bytes()

	//一段时间后，签署一个消息（hashed to h, as h^x）
	message := "alice's some text to sign"
	h := pairing.NewG1().SetFromStringHash(message, sha256.New())
	signature := pairing.NewG2().PowZn(h, privKey)

	//发送消息和签名给bob
	messageChannel <- &messageData{message: message, signature: signature.Bytes()}
	finished <- true
}

//bob验证从alice那接收到的消息
func bob(sharedParams string, sharedG []byte, messageChannel chan *messageData, keyChannel chan []byte, finished chan bool) {
	//bob加载系统公共参数
	pairing, _ := pbc.NewPairingFromString(sharedParams)
	g := pairing.NewG2().SetBytes(sharedG)

	//bob接收alice的公钥（假如手动校验）
	pubKey := pairing.NewG2().SetBytes(<-keyChannel)

	//一段时间后，bob收到了一条消息来验证
	data := <-messageChannel
	signature := pairing.NewG1().SetBytes(data.signature)

	//校验, Bob检查 e(h,g^x)=e(sig,g)
	h := pairing.NewG1().SetFromStringHash(data.message, sha256.New())
	temp1 := pairing.NewGT().Pair(h, pubKey)
	temp2 := pairing.NewGT().Pair(signature, g)
	if !temp1.Equals(temp2) {
		fmt.Println("*PBC* Signature check failed(bob verity alice)")
	} else {
		fmt.Println("Signature verified correctly(bob verity alice)")
	}
	finished <- true
}
