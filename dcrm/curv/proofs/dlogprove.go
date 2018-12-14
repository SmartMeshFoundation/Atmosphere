package proofs

import (
	"math/big"

	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/secret_sharing"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var S = secp256k1.S256()

type DLogProof struct {
	PK                *secret_sharing.GE
	pkTRandCommitment *secret_sharing.GE
	ChallengeResponse *big.Int
}

func (d *DLogProof) String() string {
	return fmt.Sprintf("dlog={pk=%s,pkt=%s,challengeresponse=%s}",
		secret_sharing.Xytostr(d.PK.X, d.PK.Y),
		secret_sharing.Xytostr(d.pkTRandCommitment.X, d.pkTRandCommitment.Y),
		d.ChallengeResponse.Text(16),
	)
}
func Prove(sk *big.Int) *DLogProof {
	key, _ := crypto.GenerateKey()
	key.D = big.NewInt(37)
	skTRandCommitment := key.D
	pk_t_rand_commitment_x, pk_t_rand_commitment_y := secp256k1.S256().ScalarBaseMult(skTRandCommitment.Bytes())
	pkx, pky := crypto.S256().ScalarBaseMult(sk.Bytes())
	challenge := utils.Sha3(pk_t_rand_commitment_x.Bytes(),
		secp256k1.S256().Gx.Bytes(),
		pkx.Bytes())
	challengeSK := new(big.Int).SetBytes(challenge[:])
	log.Trace(fmt.Sprintf("challengeSK=%s", challengeSK.Text(16)))
	challengeSK.Mod(challengeSK, S.N)
	secret_sharing.ModMul(challengeSK, sk)
	challengeResponse := secret_sharing.ModSub(skTRandCommitment, challengeSK)
	return &DLogProof{
		PK:                &secret_sharing.GE{pkx, pky},
		pkTRandCommitment: &secret_sharing.GE{pk_t_rand_commitment_x, pk_t_rand_commitment_y},
		ChallengeResponse: challengeResponse,
	}
}

//不会修改任何proof的内容
func Verify(proof *DLogProof) bool {
	challenge := utils.Sha3(
		proof.pkTRandCommitment.X.Bytes(),
		S.Gx.Bytes(),
		proof.PK.X.Bytes(),
	)
	challengeSK := new(big.Int).SetBytes(challenge[:])
	challengeSK.Mod(challengeSK, S.N)
	pkChallengeX, pkChallengeY := S.ScalarMult(proof.PK.X, proof.PK.Y, challengeSK.Bytes())
	pkVerifierX, pkVerifierY := S.ScalarBaseMult(proof.ChallengeResponse.Bytes())
	pkVerifierX, pkVerifierY = secret_sharing.PointAdd(pkVerifierX, pkVerifierY, pkChallengeX, pkChallengeY)
	return pkVerifierX.Cmp(proof.pkTRandCommitment.X) == 0 &&
		pkVerifierY.Cmp(proof.pkTRandCommitment.Y) == 0
}
