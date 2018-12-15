package proofs

import (
	"math/big"

	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/share"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var S = secp256k1.S256()

//证明Pk这个公钥对应的私钥,我有
type DLogProof struct {
	PK                *share.SPubKey
	pkTRandCommitment *share.SPubKey
	ChallengeResponse share.SPrivKey
}

func (d *DLogProof) String() string {
	return fmt.Sprintf("dlog={pk=%s,pkt=%s,challengeresponse=%s}",
		share.Xytostr(d.PK.X, d.PK.Y),
		share.Xytostr(d.pkTRandCommitment.X, d.pkTRandCommitment.Y),
		d.ChallengeResponse,
	)
}
func Prove(sk share.SPrivKey) *DLogProof {
	//todo fixme bai
	//key.D = big.NewInt(37)
	skTRandCommitment := share.RandomPrivateKey()
	randCommitmentX, randCommitmentY := secp256k1.S256().ScalarBaseMult(skTRandCommitment.Bytes())
	pkx, pky := crypto.S256().ScalarBaseMult(sk.D.Bytes())
	challenge := utils.ShaSecret(randCommitmentX.Bytes(),
		secp256k1.S256().Gx.Bytes(),
		pkx.Bytes())
	challengeSK := share.BigInt2PrivateKey(new(big.Int).SetBytes(challenge[:]))
	//log.Trace(fmt.Sprintf("challengeSK=%s", challengeSK))
	//challengeSK.Mod(challengeSK, S.N)
	//log.Trace(fmt.Sprintf("challenge_fe=%s", challengeSK))
	share.ModMul(challengeSK, sk)

	challengeResponse := share.ModSub(skTRandCommitment, challengeSK)
	return &DLogProof{
		PK:                &share.SPubKey{pkx, pky},
		pkTRandCommitment: &share.SPubKey{randCommitmentX, randCommitmentY},
		ChallengeResponse: challengeResponse,
	}
}

//不会修改任何proof的内容 const
func Verify(proof *DLogProof) bool {
	challenge := utils.ShaSecret(
		proof.pkTRandCommitment.X.Bytes(),
		S.Gx.Bytes(),
		proof.PK.X.Bytes(),
	)
	challengeSK := new(big.Int).SetBytes(challenge[:])
	challengeSK.Mod(challengeSK, S.N)
	pkChallengeX, pkChallengeY := S.ScalarMult(proof.PK.X, proof.PK.Y, challengeSK.Bytes())
	pkVerifierX, pkVerifierY := S.ScalarBaseMult(proof.ChallengeResponse.Bytes())
	pkVerifierX, pkVerifierY = share.PointAdd(pkVerifierX, pkVerifierY, pkChallengeX, pkChallengeY)
	return pkVerifierX.Cmp(proof.pkTRandCommitment.X) == 0 &&
		pkVerifierY.Cmp(proof.pkTRandCommitment.Y) == 0
}
