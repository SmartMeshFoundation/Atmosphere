package commitments

import (
	"crypto/sha256"
	"math/big"
	"math/rand"

	"github.com/Nik-U/pbc"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/sirupsen/logrus"
)

type MultiTrapdoorCommitment struct {
	Commitment *Commitment
	Open       *Open
}

func (mtdc *MultiTrapdoorCommitment) Constructor(commitment *Commitment, open *Open) {
	mtdc.Commitment = commitment
	mtdc.Open = open
}

func MultiLinnearCommit(rand *rand.Rand, mpk *MultiTrapdoorMasterPublicKey, secrets []*big.Int) *MultiTrapdoorCommitment {
	e := mpk.pairing.NewZr()
	e.Rand()
	r := mpk.pairing.NewZr()
	r.Rand()

	h := func(target *pbc.Element, megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}
	//BigInteger digest = new BigInteger(Util.sha256Hash(secretsBytes)).mod(mpk.q); // AR mod
	secretsBytes := make([]string, len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen() + 7) / 8)
		se := make([]byte, count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}
	digest := mpk.pairing.NewZr()
	h(digest, secretsBytes[:])

	// Point he = curve.add(mpk.h, curve.multiply(mpk.g, new BigInt(e)));he=h+(g*e)
	gMule := mpk.pairing.NewG1()
	gMule = gMule.MulZn(mpk.g, e)
	he := mpk.pairing.NewG1()
	he = he.Add(mpk.h, gMule)
	// Point a = curve.add(curve.multiply(mpk.g, new BigInt(digest)), curve.multiply(he, new BigInt(r)));
	dMulg := mpk.pairing.NewG1()
	dMulg = dMulg.MulZn(mpk.g, digest)
	heMulr := mpk.pairing.NewG1()
	heMulr = heMulr.MulZn(he, r)
	a := mpk.pairing.NewG1()
	a = a.Add(dMulg, heMulr)

	open := new(Open)
	open.Constructor(r, secrets)
	commitment := new(Commitment)
	commitment.Constructor(e, a)

	mtdct := new(MultiTrapdoorCommitment)
	mtdct.Constructor(commitment, open)
	return mtdct
	/*e := mpk.pairing.NewZr()
	e.Rand()
	r := mpk.pairing.NewZr()
	r.Rand()

	h := func(target *pbc.Element, megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}

	secretsBytes := make([]string, len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen() + 7) / 8)
		se := make([]byte, count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}

	digest := mpk.pairing.NewZr()
	h(digest, secretsBytes[:])

	ge := mpk.pairing.NewG1()
	ge.MulZn(mpk.g, e)

	he := mpk.pairing.NewG1()
	he.Add(mpk.h, ge)

	rhe := mpk.pairing.NewG1()
	rhe.MulZn(he, r)

	dg := mpk.pairing.NewG1()
	dg.MulZn(mpk.g, digest)

	a := mpk.pairing.NewG1()
	a.Add(dg, rhe)

	open := new(Open)
	open.Constructor(r, secrets)
	commitment := new(Commitment)
	commitment.Constructor(e, a)

	mtdct := new(MultiTrapdoorCommitment)
	mtdct.Constructor(commitment, open)

	return mtdct*/
}

func CheckCommitment(commitment *Commitment, open *Open, mpk *MultiTrapdoorMasterPublicKey) bool {
	g := mpk.g
	h := mpk.h
	f := func(target *pbc.Element, megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}
	secrets := open.GetSecrets()
	secretsBytes := make([]string, len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen() + 7) / 8)
		se := make([]byte, count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}
	// digest hash
	digest := mpk.pairing.NewZr()
	f(digest, secretsBytes[:])
	// a=curve.multiply(g,new BigInt(open.getRandomness()))
	a := mpk.pairing.NewG1()
	a = a.MulZn(g, open.getRandomness())
	// b=curve.add(h, curve.multiply(g, new BigInt(commitment.pubkey)))
	gMulp := mpk.pairing.NewG1()
	gMulp = gMulp.MulZn(g, commitment.pubkey)
	b := mpk.pairing.NewG1()
	b = b.Add(h, gMulp)
	// c=curve.add(commitment.committment, curve.multiply(g, new BigInt(digest.negate())))
	gMulneg := mpk.pairing.NewG1()
	digest = digest.Neg(digest)
	gMulneg = gMulneg.MulZn(g, digest)
	c := mpk.pairing.NewG1()
	c = c.Add(commitment.committment, gMulneg)

	result := DDHTest(a, b, c, g, mpk.pairing)
	if result == false {
		logrus.Error("Verify commitment failed")
	}
	return result
	/*g := mpk.g
	h := mpk.h

	f := func(target *pbc.Element, megs []string) {
		hash := sha256.New()
		for j := range megs {
			hash.Write([]byte(megs[j]))
		}
		i := &big.Int{}
		target.SetBig(i.SetBytes(hash.Sum([]byte{})))
	}

	secrets := open.GetSecrets()
	secretsBytes := make([]string, len(secrets))
	for i := range secrets {
		count := ((secrets[i].BitLen() + 7) / 8)
		se := make([]byte, count)
		math.ReadBits(secrets[i], se[:])
		secretsBytes[i] = string(se[:])
	}
	//digest hash(h秘密)
	digest := mpk.pairing.NewZr()
	f(digest, secretsBytes[:])
	//g^a
	rg := mpk.pairing.NewG1()
	rg.MulZn(g, open.getRandomness())
	//(g,h)
	d1 := mpk.pairing.NewG1()
	d1.MulZn(g, commitment.Pubkey)
	//h^b
	dh := mpk.pairing.NewG1()
	dh.Add(h, d1)
	//g(-digest)
	gdn := mpk.pairing.NewG1()
	digest.Neg(digest)
	gdn.MulZn(g, digest)
	//a*b
	comd := mpk.pairing.NewG1()
	comd.Add(commitment.Commitment, gdn)
	b := DDHTest(rg, dh, comd, g, mpk.pairing)
	if b == false {
		logrus.Error("Check commitment error")
	}
	return b*/
}

//DDHTest
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
func DDHTest(a *pbc.Element, b *pbc.Element, c *pbc.Element, generator *pbc.Element, pairing *pbc.Pairing) bool {
	temp1 := pairing.NewGT().Pair(a, b)
	temp2 := pairing.NewGT().Pair(generator, c)
	return temp1.Equals(temp2)
}
