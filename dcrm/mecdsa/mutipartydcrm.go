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

func KeyGenTN(t, n int) ([]*Keys, []*SharedKeys, []*secret_sharing.GE, *secret_sharing.GE, *secret_sharing.VerifiableSS, error) {
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
			return nil, nil, nil, nil, nil, err
		}
		vss_scheme_vec = append(vss_scheme_vec, vss_scheme)
		secret_shares_vec = append(secret_shares_vec, secret_shares)
		index_vec = append(index_vec, index)
	}
	var vss_scheme_for_test secret_sharing.VerifiableSS
	vss_scheme_for_test = *vss_scheme_vec[0]
	//log.Trace(fmt.Sprintf("secret_shares_vec=%s", utils.StringInterface(secret_shares_vec, 5)))
	var partyShares [][]*big.Int
	for i := 0; i < n; i++ {
		var s []*big.Int
		for j := 0; j < n; j++ {
			s = append(s, secret_shares_vec[j][i])
		}
		partyShares = append(partyShares, s)
	}
	//log.Trace(fmt.Sprintf("partyShares=%s", utils.StringInterface(partyShares, 5)))

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
			return nil, nil, nil, nil, nil, err
		}
		shared_keys_vec = append(shared_keys_vec, shared_keys)
		dlog_proof_vec = append(dlog_proof_vec, dlog_proof)

	}

	if !VerifyDlogProofs(parames, dlog_proof_vec, y_vec) {
		return nil, nil, nil, nil, nil, errors.New("bad dlog proof")
	}
	//t = n - 1
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
		return nil, nil, nil, nil, nil, errors.New("sum not equal")
	}
	var pk_vec []*secret_sharing.GE
	for i := 0; i < len(dlog_proof_vec); i++ {
		pk_vec = append(pk_vec, dlog_proof_vec[i].PK)
	}
	return party_keys_vec, shared_keys_vec, pk_vec, sumy, vss_scheme_vec[0], nil
}

func Sign(t, n, ttag int, s []int) error {

	party_keys_vec, shared_keys_vec, _, y, vss_scheme, err := KeyGenTN(t, n)
	if err != nil {
		return err
	}
	if ttag <= t {
		return errors.New("ttag  <=t ")
	}
	// each party creates a signing key. This happens in parallel IRL. In this test we
	// create a vector of signing keys, one for each party.
	// throughout i will index parties
	var sign_keys_vec []*SignKeys
	for i := 0; i < ttag; i++ {
		k := createSignKeys(shared_keys_vec[s[i]], vss_scheme, s[i], s)
		sign_keys_vec = append(sign_keys_vec, k)
	}

	// each party computes [Ci,Di] = com(g^gamma_i) and broadcast the commitmentsvar
	var bc1_vec []*SignBroadcastPhase1
	var blind_vec1 []*big.Int
	for i := 0; i < ttag; i++ {
		com, blind := sign_keys_vec[i].phase1Broadcast()
		bc1_vec = append(bc1_vec, com)
		blind_vec1 = append(blind_vec1, blind)
	}

	// each party i sends encryption of k_i under her Paillier key
	// m_a_vec = [ma_0;ma_1;,...]
	var m_a_vec []*MessageA
	for i := 0; i < ttag; i++ {
		m_a, err := NewMessageA(sign_keys_vec[i].ki, &party_keys_vec[s[i]].dk.PublicKey)
		if err != nil {
			return err
		}
		m_a_vec = append(m_a_vec, m_a)
	}

	// each party i sends responses to m_a_vec she received (one response with input gamma_i and one with w_i)
	// m_b_gamma_vec_all is a matrix where column i is a vector of message_b's that party i answers to all ma_{j!=i} using paillier key of party j to answer to ma_j

	// aggregation of the n messages of all parties
	var m_b_gamma_vec_all [][]*MessageB
	var beta_vec_all [][]*big.Int
	var m_b_w_vec_all [][]*MessageB
	var ni_vec_all [][]*big.Int

	for i := 0; i < ttag; i++ {
		var m_b_gamma_vec []*MessageB
		var beta_vec []*big.Int
		var m_b_w_vec []*MessageB
		var ni_vec []*big.Int
		for j := 0; j < ttag-1; j++ {
			ind := j
			if j >= i {
				ind = j + 1
			}
			mbGamma, betaGamma, err := NewMessageB(
				sign_keys_vec[i].gammaI,
				&party_keys_vec[s[ind]].dk.PublicKey,
				m_a_vec[ind])
			if err != nil {
				return err
			}
			mbw, betawi, err := NewMessageB(
				sign_keys_vec[i].wi,
				&party_keys_vec[s[ind]].dk.PublicKey,
				m_a_vec[ind],
			)
			if err != nil {
				return err
			}
			m_b_gamma_vec = append(m_b_gamma_vec, mbGamma)
			beta_vec = append(beta_vec, betaGamma)
			m_b_w_vec = append(m_b_w_vec, mbw)
			ni_vec = append(ni_vec, betawi)
		}
		m_b_gamma_vec_all = append(m_b_gamma_vec_all, m_b_gamma_vec)
		beta_vec_all = append(beta_vec_all, beta_vec)
		m_b_w_vec_all = append(m_b_w_vec_all, m_b_w_vec)
		ni_vec_all = append(ni_vec_all, ni_vec)
	}

	// Here we complete the MwA protocols by taking the mb matrices and starting with the first column generating the appropriate message
	// for example for index i=0 j=0 we need party at index s[1] to answer to mb that party s[0] sent, completing a protocol between s[0] and s[1].
	//  for index i=1 j=0 we need party at index s[0] to answer to mb that party s[1]. etc.
	// IRL each party i should get only the mb messages that other parties sent in response to the party i ma's.
	// TODO: simulate as IRL
	var alpha_vec_all [][]*big.Int
	var miu_vec_all [][]*big.Int
	for i := 0; i < ttag; i++ {
		var alpha_vec []*big.Int
		var miu_vec []*big.Int

		m_b_gamma_vec_i := m_b_gamma_vec_all[i]
		m_b_w_vec_i := m_b_gamma_vec_all[i]
		for j := 0; j < ttag-1; j++ {
			ind := j
			if j >= i {
				ind = j + 1
			}
			mb := m_b_gamma_vec_i[j]
			alpha_ij_gamma, err := mb.VerifyProofsGetAlpha(
				party_keys_vec[s[ind]].dk,
				sign_keys_vec[ind].ki,
			)
			if err != nil {
				return fmt.Errorf("wrong dlog or m_b %s", err)
			}
			mb = m_b_w_vec_i[j]
			alpha_ij_wi, err := mb.VerifyProofsGetAlpha(
				party_keys_vec[s[ind]].dk,
				sign_keys_vec[ind].ki,
			)
			if err != nil {
				return fmt.Errorf("wrong dlog or m_b2  %s", err)
			}

			// since we actually run two MtAwc each party needs to make sure that the values B are the same as the public values
			// here for b=w_i the parties already know W_i = g^w_i  for each party so this check is done here. for b = gamma_i the check will be later when g^gamma_i will become public
			// currently we take the W_i from the other parties signing keys
			// TODO: use pk_vec (first change from x_i to w_i) for this check.
			if !EqualGE(mb.BProof.PK, sign_keys_vec[i].gwi) {
				log.Trace(fmt.Sprintf("mb.BProof.PK=%s", utils.StringInterface(mb.BProof.PK, 3)))
				log.Trace(fmt.Sprintf("sign_keys_vec[i].gwi=%s", utils.StringInterface(sign_keys_vec[i].gwi, 3)))
				panic("not equal")
			}
			alpha_vec = append(alpha_vec, alpha_ij_gamma)
			miu_vec = append(miu_vec, alpha_ij_wi)
		}
		alpha_vec_all = append(alpha_vec_all, alpha_vec)
		miu_vec_all = append(miu_vec_all, miu_vec)
	}

	var delta_vec []*big.Int
	var sigma_vec []*big.Int

	for i := 0; i < ttag; i++ {
		delta := sign_keys_vec[i].phase2DeltaI(alpha_vec_all[i], beta_vec_all[i])
		sigma := sign_keys_vec[i].phase2SigmaI(miu_vec_all[i], ni_vec_all[i])
		delta_vec = append(delta_vec, delta)
		sigma_vec = append(sigma_vec, sigma)
	}

	// all parties broadcast delta_i and compute delta_i ^(-1)
	delta_inv := phase3ReconstructDelta(delta_vec)

	// de-commit to g^gamma_i from phase1, test comm correctness, and that it is the same value used in MtA.
	// Return R
	var gGammaIVec []*secret_sharing.GE
	for i := 0; i < len(sign_keys_vec); i++ {
		gGammaIVec = append(gGammaIVec, sign_keys_vec[i].gGammaI.Clone())
	}
	var r_vec []*secret_sharing.GE
	for i := 0; i < ttag; i++ {
		// each party i tests all B = g^b = g ^ gamma_i she received.
		var bproofs []*proofs.DLogProof
		for j := 0; j < ttag; j++ {
			bproofs = append(bproofs, m_b_gamma_vec_all[j][0].BProof)
		}
		r, err := phase4(delta_inv, bproofs, blind_vec1, gGammaIVec, bc1_vec)
		if err != nil {
			return fmt.Errorf("bad gamma_i decommit %s", err)
		}
		r_vec = append(r_vec, r)
	}

	message := big.NewInt(399558858588)
	messageHash := utils.Sha3(message.Bytes())
	message = message.SetBytes(messageHash[:])
	var local_sig_vec []*LocalSignature
	// each party computes s_i but don't send it yet. we start with phase5

	for i := 0; i < ttag; i++ {
		localSig := phase5LocalSignature(sign_keys_vec[i].ki, message, r_vec[i], sigma_vec[i], y)
		local_sig_vec = append(local_sig_vec, localSig)
	}

	var phase5_com_vec []*Phase5Com1
	var phase5a_decom_vec []*Phase5ADecom1
	var helgamal_proof_vec []*proofs.HomoELGamalProof
	// we notice that the proof for V= R^sg^l, B = A^l is a general form of homomorphic elgamal.
	for i := 0; i < ttag; i++ {
		phase5Com, phase5ADecom, helgamalProof := local_sig_vec[i].phase5aBroadcast5bZkproof()
		phase5_com_vec = append(phase5_com_vec, phase5Com)
		phase5a_decom_vec = append(phase5a_decom_vec, phase5ADecom)
		helgamal_proof_vec = append(helgamal_proof_vec, helgamalProof)
	}

	var phase5_com2_vec []*Phase5Com2
	var phase_5d_decom2_vec []*Phase5DDecom2
	for i := 0; i < ttag; i++ {
		var decom_i []*Phase5ADecom1
		var com_i []*Phase5Com1
		var elgamla_i []*proofs.HomoELGamalProof

		decom_i = append(decom_i, phase5a_decom_vec[:i]...)
		decom_i = append(decom_i, phase5a_decom_vec[i+1:]...)

		com_i = append(com_i, phase5_com_vec[:i]...)
		com_i = append(com_i, phase5_com_vec[i+1:]...)

		elgamla_i = append(elgamla_i, helgamal_proof_vec[:i]...)
		elgamla_i = append(elgamla_i, helgamal_proof_vec[i+1:]...)

		phase5com2, phase5ddecom2, err := local_sig_vec[i].phase5c(
			phase5a_decom_vec, phase5_com_vec,
			helgamal_proof_vec, phase5a_decom_vec[i].vi, r_vec[0],
		)
		if err != nil {
			return fmt.Errorf("error phase5 err %s", err)
		}
		phase5_com2_vec = append(phase5_com2_vec, phase5com2)
		phase_5d_decom2_vec = append(phase_5d_decom2_vec, phase5ddecom2)
	}

	// assuming phase5 checks passes each party sends s_i and compute sum_i{s_i}
	var s_vec []*big.Int
	for i := 0; i < ttag; i++ {
		si, err := local_sig_vec[i].phase5d(phase_5d_decom2_vec, phase5_com2_vec, phase5a_decom_vec)
		if err != nil {
			return fmt.Errorf("bad com 5d %s", err)
		}
		s_vec = append(s_vec, si)
	}
	// here we compute the signature only of party i=0 to demonstrate correctness.
	_, _, err = local_sig_vec[0].outputSignature(s_vec[1:])
	if err != nil {
		return fmt.Errorf("sigature verify  err %s", err)
	}
	return nil
}
