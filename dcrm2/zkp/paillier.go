package zkp

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
)

var one = big.NewInt(1)
var ErrMessageTooLong = errors.New("paillier: message too long for Paillier public key size")

//生成RSA私钥
func GenerateKey(random io.Reader, bits int) (*PrivateKey, error) {
	p, err := rand.Prime(random, bits)
	if err != nil {
		return nil, err
	}

	q, err := rand.Prime(random, bits)
	if err != nil {
		return nil, err
	}

	// n = p * q
	n := new(big.Int).Mul(p, q)

	// l = phi(n) = (p-1) * q(-1)
	l := new(big.Int).Mul(
		new(big.Int).Sub(p, one),
		new(big.Int).Sub(q, one),
	)

	return &PrivateKey{
		PublicKey: PublicKey{
			N:        n,
			NSquared: new(big.Int).Mul(n, n),
			G:        new(big.Int).Add(n, one), // g = n + 1
		},
		L: l,
		U: new(big.Int).ModInverse(l, n),
	}, nil
}

type PrivateKey struct {
	PublicKey
	L *big.Int // phi(n), (p-1)*(q-1)
	U *big.Int // l^-1 mod n
}

type PublicKey struct {
	N        *big.Int // modulus
	G        *big.Int // n+1, since p and q are same length
	NSquared *big.Int
}

//pubKey 公钥
// m 明码,私钥片明文
// r 计算参数
func Encrypt(pubKey *PublicKey, m *big.Int, r *big.Int) *big.Int {
	s, _ := encrypt(pubKey, m.Bytes(), r)
	return new(big.Int).SetBytes(s)
}

func Decrypt(privKey *PrivateKey, c *big.Int) *big.Int {
	s, _ := decrypt(privKey, c.Bytes())
	return new(big.Int).SetBytes(s)
}
func CipherAdd(pubKey *PublicKey, c1 *big.Int, c2 *big.Int) *big.Int {
	s := addCipher(pubKey, c1.Bytes(), c2.Bytes())
	return new(big.Int).SetBytes(s)
}
func CipherMultiply(pubKey *PublicKey, c1 *big.Int, cons *big.Int) *big.Int {
	s := Mul(pubKey, c1.Bytes(), cons.Bytes())
	return new(big.Int).SetBytes(s)
}

// c = g^m * r^n mod n^2
//pubKey只是选定的一个公共已知点,是计算的参数,并不是用来签名.因此就算是其他公证人知道pubkey对应的私钥,也不能解密得到plainText.
func encrypt(pubKey *PublicKey, plainText []byte, r *big.Int) ([]byte, error) {
	m := new(big.Int).SetBytes(plainText)
	if pubKey.N.Cmp(m) < 1 { // N < m
		return nil, ErrMessageTooLong
	}

	n := pubKey.N
	c := new(big.Int).Mod(
		new(big.Int).Mul(
			new(big.Int).Exp(pubKey.G, m, pubKey.NSquared),
			new(big.Int).Exp(r, n, pubKey.NSquared),
		),
		pubKey.NSquared,
	)

	return c.Bytes(), nil
}

// decrypt decrypts the passed cipher text.
func decrypt(privKey *PrivateKey, cipherText []byte) ([]byte, error) {
	c := new(big.Int).SetBytes(cipherText)
	if privKey.NSquared.Cmp(c) < 1 { // c < n^2
		return nil, ErrMessageTooLong
	}

	// c^l mod n^2
	a := new(big.Int).Exp(c, privKey.L, privKey.NSquared)

	// L(a)
	// (a - 1) / n
	l := new(big.Int).Div(
		new(big.Int).Sub(a, one),
		privKey.N,
	)

	// m = L(c^l mod n^2) * u mod n
	m := new(big.Int).Mod(
		new(big.Int).Mul(l, privKey.U),
		privKey.N,
	)

	return m.Bytes(), nil
}

// addCipher homomorphically adds together two cipher texts.
// To do this we multiply the two cipher texts, upon decryption, the resulting
// plain text will be the sum of the corresponding plain texts.
func addCipher(pubKey *PublicKey, cipher1, cipher2 []byte) []byte {
	x := new(big.Int).SetBytes(cipher1)
	y := new(big.Int).SetBytes(cipher2)

	// x * y mod n^2
	return new(big.Int).Mod(
		new(big.Int).Mul(x, y),
		pubKey.NSquared,
	).Bytes()
}

// Add homomorphically adds a passed constant to the encrypted integer
// (our cipher text). We do this by multiplying the constant with our
// ciphertext. Upon decryption, the resulting plain text will be the sum of
// the plaintext integer and the constant.
func Add(pubKey *PublicKey, cipher, constant []byte) []byte {
	c := new(big.Int).SetBytes(cipher)
	x := new(big.Int).SetBytes(constant)

	// c * g ^ x mod n^2
	return new(big.Int).Mod(
		new(big.Int).Mul(c, new(big.Int).Exp(pubKey.G, x, pubKey.NSquared)),
		pubKey.NSquared,
	).Bytes()
}

// Mul homomorphically multiplies an encrypted integer (cipher text) by a
// constant. We do this by raising our cipher text to the power of the passed
// constant. Upon decryption, the resulting plain text will be the product of
// the plaintext integer and the constant.
func Mul(pubKey *PublicKey, cipher []byte, constant []byte) []byte {
	c := new(big.Int).SetBytes(cipher)
	x := new(big.Int).SetBytes(constant)

	// c ^ x mod n^2
	return new(big.Int).Exp(c, x, pubKey.NSquared).Bytes()
}
