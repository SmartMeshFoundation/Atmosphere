package kgcenter

import (
	"fmt"
	"math/rand"
	"math/big"
	"crypto/sha256"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/Nik-U/pbc"
	"github.com/sirupsen/logrus"
)
type Commitment struct {
	committment *pbc.Element
	pubkey *pbc.Element
}

//MultiTrapdoorCommitment
type MultiTrapdoorCommitment struct {
	commitment *Commitment
	open *Open
}

type Open struct {
	secrets []*big.Int
	randomness *pbc.Element
}

type CmtMasterPublicKey struct {
	g *pbc.Element
	q *big.Int
	h *pbc.Element
	pairing *pbc.Pairing
}

func (ct *Commitment) New(pubkey *pbc.Element,a *pbc.Element) {
	ct.pubkey = pubkey
	ct.committment = a
}

func (open *Open) New(randomness *pbc.Element,secrets []*big.Int) {
	open.randomness = randomness

	open.secrets = secrets //test
}

func (open *Open) GetSecrets() []*big.Int {
	return open.secrets
}

func (open *Open) getRandomness() *pbc.Element {
	return open.randomness
}


func (mtdct *MultiTrapdoorCommitment) New(commitment *Commitment,open *Open) {
	mtdct.commitment = commitment
	mtdct.open = open
}

func (cmpk *CmtMasterPublicKey) New(g *pbc.Element,q *big.Int,h *pbc.Element,pairing *pbc.Pairing) {
	cmpk.g = g
	cmpk.q = q
	cmpk.h = h
	cmpk.pairing = pairing
}

/*
public static MultiTrapdoorCommitment multilinnearCommit(Random rand,
			MultiTrapdoorMasterPublicKey mpk, BigInteger... secrets) {
		EllipticCurve curve = mpk.pairing.getCurve();
		BigInteger e = Util.randomFromZn(mpk.q, rand);
		BigInteger r = Util.randomFromZn(mpk.q, rand);
		byte[][] secretsBytes = new byte[secrets.length][];
		for (int i = 0; i < secrets.length; i++) {
			secretsBytes[i] = secrets[i].toByteArray();
		}
		BigInteger digest = new BigInteger(Util.sha256Hash(secretsBytes))
				.mod(mpk.q); // AR mod
		Point he = curve.add(mpk.h, curve.multiply(mpk.g, new BigInt(e)));
		Point a = curve.add(curve.multiply(mpk.g, new BigInt(digest)), curve.multiply(he, new BigInt(r)));
		Open<BigInteger> open = new Open<BigInteger>(r, secrets);
		Commitment commitment = new Commitment(e, a);

		return new MultiTrapdoorCommitment(commitment, open);

}*/
func MultiLinnearCommit(rnd *rand.Rand,mpk *CmtMasterPublicKey,secrets []*big.Int) *MultiTrapdoorCommitment {
	e := mpk.pairing.NewZr()
	e.Rand()
	r := mpk.pairing.NewZr()
	r.Rand()

	h := func(target *pbc.Element,megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}

	secretsBytes := make([]string,len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen()+7)/8)
		se := make([]byte,count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}

	digest := mpk.pairing.NewZr()
	h(digest,secretsBytes[:])

	ge := mpk.pairing.NewG1()
	ge.MulZn(mpk.g,e)

	//he = mpk.h + ge
	he := mpk.pairing.NewG1()
	he.Add(mpk.h,ge)

	//he = r*he
	rhe := mpk.pairing.NewG1()
	rhe.MulZn(he,r)

	//dg = digest*mpk.g
	dg := mpk.pairing.NewG1()
	//
	dg.MulZn(mpk.g,digest)

	//a = mpk.g + he
	a := mpk.pairing.NewG1()
	a.Add(dg,rhe)

	open := new(Open)
	open.New(r,secrets)
	commitment := new(Commitment)
	commitment.New(e,a)

	mtdct := new(MultiTrapdoorCommitment)
	mtdct.New(commitment,open)

	return mtdct
}

//初始化配对pairing_init_set_str 0成功1失败
func GenerateMasterPK() *CmtMasterPublicKey {
	pairing, err := pbc.NewPairingFromString("type a\nq 7313295762564678553220399414112155363840682896273128302543102778210584118101444624864132462285921835023839111762785054210425140241018649354445745491039387\nh 10007920040268628970387373215664582404186858178692152430205359413268619141100079249246263148037326528074908\nr 730750818665451459101842416358141509827966402561\nexp2 159\nexp1 17\nsign1 1\nsign0 1\n")
	if err != nil {
		fmt.Println("preload pairing fail.\n")
	}

	g := getBasePoint(pairing)
	q,_ := new(big.Int).SetString("730750818665451459101842416358141509827966402561",10)
	h := RandomPointInG1(pairing)
	cmpk := new(CmtMasterPublicKey)
	cmpk.New(g,q,h,pairing)
	return cmpk
}

func RandomPointInG1(pairing *pbc.Pairing) *pbc.Element {
	for{
		h := pairing.NewG1()
		h.Rand()

		cof := pairing.NewZr()
		num,_ := new(big.Int).SetString("10007920040268628970387373215664582404186858178692152430205359413268619141100079249246263148037326528074908",10)
		cof.SetBig(num)

		hh := pairing.NewG1()
		hh.MulZn(h,cof)

		order,_ := new(big.Int).SetString("730750818665451459101842416358141509827966402561",10)
		q := pairing.NewZr()
		q.SetBig(order)

		hhh := pairing.NewG1()
		hhh.MulZn(hh,q)

		if hhh.Is0() {
			return hh
		}
	}
	return nil
}

func getBasePoint(pairing *pbc.Pairing) *pbc.Element {
	var p *pbc.Element
	cof := pairing.NewZr()
	num,_ := new(big.Int).SetString("10007920040268628970387373215664582404186858178692152430205359413268619141100079249246263148037326528074908",10)
	cof.SetBig(num)

	order,_ := new(big.Int).SetString("730750818665451459101842416358141509827966402561",10)
	q := pairing.NewZr()
	q.SetBig(order)

	for {
		p = pairing.NewG1()
		p.Rand()
		ge := pairing.NewG1()
		ge.MulZn(p,cof)

		pq := pairing.NewG1()
		pq.MulZn(ge,q)

		if ge.Is0() || pq.Is0() {
			return ge
		}
	}

	return nil
}

func (mtc *MultiTrapdoorCommitment) CmtOpen() *Open {
	return mtc.open
}

func (mtc *MultiTrapdoorCommitment) CmtCommitment() *Commitment {
	return mtc.commitment
}

//
func Checkcommitment(commitment *Commitment,open *Open,mpk *CmtMasterPublicKey) bool {
	g := mpk.g
	h := mpk.h

	f := func(target *pbc.Element,megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}

	secrets := open.GetSecrets();
	secretsBytes := make([]string,len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen()+7)/8)
		se := make([]byte,count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}
	//digest hash(h秘密)
	digest := mpk.pairing.NewZr()
	f(digest,secretsBytes[:])
	//g^a
	rg := mpk.pairing.NewG1()
	rg.MulZn(g,open.getRandomness())
	//(g,h)
	d1 := mpk.pairing.NewG1()
	d1.MulZn(g,commitment.pubkey)
	//h^b
	dh := mpk.pairing.NewG1()
	dh.Add(h,d1)
	//g(-digest)
	gdn := mpk.pairing.NewG1()
	digest.Neg(digest)
	gdn.MulZn(g,digest)
	//a*b
	comd := mpk.pairing.NewG1()
	comd.Add(commitment.committment,gdn)
	b := DDHTest(rg,dh,comd,g,mpk.pairing)
	if b == false {
		logrus.Error("Check commitment error")
	}
	return b
}

//见 main2 pairing(g,h)^(a*b)
//a=generator
func DDHTest(a *pbc.Element,b *pbc.Element,c *pbc.Element,generator *pbc.Element,pairing *pbc.Pairing) bool {
	/*temp1 := pairing.NewGT().Pair(h, pubKey)
	temp2 := pairing.NewGT().Pair(signature, g)*/
	temp1 := pairing.NewGT().Pair(a, b)
	temp2 := pairing.NewGT().Pair(generator,c)

	return temp1.Equals(temp2)//temp1=temp2
}//f(x,y)=f(x)+f(y)
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