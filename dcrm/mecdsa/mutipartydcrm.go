package mecdsa

import (
	"math/big"

	"errors"

	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/proofs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/curv/secret_sharing"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
)

func KeyGenTN(t, n int) ([]*Keys, []*SharedKeys, []*secret_sharing.GE, *secret_sharing.VerifiableSS, error) {
	parames := &Parameters{
		Threshold:  t,
		ShareCount: n,
	}
	var party_keys_vec = []*Keys{}
	//paillier keys
	for i := 0; i < n; i++ {
		party_keys_vec = append(party_keys_vec, createKeys(i))
	}

	var broadcastMsgs1 []*KeyGenBroadcastMessage1
	var blind_vec []*big.Int
	for i := 0; i < n; i++ {
		partyBroadcastMsg, partyBlind := party_keys_vec[i].phase1BroadcastPhase3ProofOfCorrectKey()
		broadcastMsgs1 = append(broadcastMsgs1, partyBroadcastMsg)
		blind_vec = append(blind_vec, partyBlind)
	}
	//---------------------------------------------------
	var y_vec []*secret_sharing.GE
	for i := 0; i < len(party_keys_vec); i++ {
		y_vec = append(y_vec, party_keys_vec[i].yi)
	}
	//公钥之和
	sumy := y_vec[0].Clone()
	for i := 1; i < len(y_vec); i++ {
		secret_sharing.PointAdd(sumy.X, sumy.Y, y_vec[i].X, y_vec[i].Y)
	}
	//shamir 配多项式
	var vss_scheme_vec []*secret_sharing.VerifiableSS
	var secret_shares_vec [][]*big.Int
	var index_vec []int
	for i := 0; i < n; i++ {
		vss_scheme, secret_shares, index, err := party_keys_vec[i].phase1VerifyComPhase3VerifyCorrectKeyPhase2Distribute(
			parames, blind_vec, y_vec, broadcastMsgs1)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		vss_scheme_vec = append(vss_scheme_vec, vss_scheme)
		secret_shares_vec = append(secret_shares_vec, secret_shares)
		index_vec = append(index_vec, index)
	}
	var vss_scheme_for_test secret_sharing.VerifiableSS
	vss_scheme_for_test = *vss_scheme_vec[0]
	log.Trace("secret_shares_vec=%s", utils.StringInterface(secret_shares_vec, 5))
	var partyShares [][]*big.Int
	for i := 0; i < n; i++ {
		var s []*big.Int
		for j := 0; j < n; j++ {
			s = append(s, secret_shares_vec[j][i])
		}
		partyShares = append(partyShares, s)
	}
	log.Trace(fmt.Sprintf("partyShares=%s", utils.StringInterface(partyShares, 5)))

	var shared_keys_vec []*SharedKeys
	var dlog_proof_vec []*proofs.DLogProof
	for i := 0; i < n; i++ {
		shared_keys, dlog_proof, err := party_keys_vec[i].phase2_verify_vss_construct_keypair_phase3_pok_dlog(parames,
			y_vec,
			partyShares[i],
			vss_scheme_vec,
			index_vec[i]+1,
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		shared_keys_vec = append(shared_keys_vec, shared_keys)
		dlog_proof_vec = append(dlog_proof_vec, dlog_proof)

	}

	if !VerifyDlogProofs(parames, dlog_proof_vec, y_vec) {
		return nil, nil, nil, nil, errors.New("bad dlog proof")
	}
	var xi []*big.Int
	for i := 0; i < t+1; i++ {
		xi = append(xi, shared_keys_vec[i].xi)
	}
	x := vss_scheme_for_test.Reconstruct(index_vec[0:t+1], xi)
	sum := big.NewInt(0)
	for i := 0; i < len(party_keys_vec); i++ {
		secret_sharing.ModAdd(sum, party_keys_vec[i].ui)
	}
	if x.Cmp(sum) != 0 {
		log.Trace(fmt.Sprintf("x=%s", x.Text(16)))
		log.Trace(fmt.Sprintf("su=%s", sum.Text(16)))
		return nil, nil, nil, nil, errors.New("sum not equal")
	}
	var pk_vec []*secret_sharing.GE
	for i := 0; i < len(dlog_proof_vec); i++ {
		pk_vec = append(pk_vec, dlog_proof_vec[i].PK)
	}
	return party_keys_vec, shared_keys_vec, pk_vec, vss_scheme_vec[0], nil
}
