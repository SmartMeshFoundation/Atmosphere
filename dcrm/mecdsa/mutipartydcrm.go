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

func KeyGenTN(t, n int) (party_keys_vec []*Keys, shared_keys_vec []*SharedKeys, pk_vec []*secret_sharing.GE, total_pubkey *secret_sharing.GE, vss0 *secret_sharing.VerifiableSS, err error) {
	parames := &Parameters{
		Threshold:  t,
		ShareCount: n,
	}
	//paillier keys
	for i := 0; i < n; i++ {
		party_keys_vec = append(party_keys_vec, createKeys(i))
	}
	log.Trace(fmt.Sprintf("party_keys_vec init=%s", utils.StringInterface(party_keys_vec, 5)))
	var broadcastMsgs1 []*KeyGenBroadcastMessage1
	var blind_vec []*big.Int
	for i := 0; i < n; i++ {
		partyBroadcastMsg, partyBlind := party_keys_vec[i].phase1BroadcastPhase3ProofOfCorrectKey()
		broadcastMsgs1 = append(broadcastMsgs1, partyBroadcastMsg)
		blind_vec = append(blind_vec, partyBlind)
	}
	//---------------------------------------------------
	//todo 私钥片对应的公钥信息是什么时候广播出去的呢?后面反复在用
	var y_vec []*secret_sharing.GE //私钥片对应的公钥集合
	for i := 0; i < len(party_keys_vec); i++ {
		y_vec = append(y_vec, party_keys_vec[i].yi)
	}
	//公钥之和
	total_pubkey = y_vec[0].Clone()
	for i := 1; i < len(y_vec); i++ {
		total_pubkey.X, total_pubkey.Y = secret_sharing.PointAdd(total_pubkey.X, total_pubkey.Y, y_vec[i].X, y_vec[i].Y)
	}
	//shamir 配多项式
	var vss_scheme_vec []*secret_sharing.VerifiableSS //vss
	var secret_shares_vec [][]*big.Int                //feldman vss提供
	var index_vec []int
	for i := 0; i < n; i++ {
		/*
				每个人都校验其他人送过来的私钥片的部分公钥,以及同态加密的公约,如果通过就使用vss机制分享自己的私钥片信息
				私钥片是可以通过Recontruct恢复的.
			vss_sheme不直接包含私钥片信息,可以直接分享给其他人.
			但是不能一次性将自己的secret_shares分享给同一个人,否则他就能恢复出来私钥片
		*/
		vss_scheme, secret_shares, index, err := party_keys_vec[i].phase1VerifyComPhase3VerifyCorrectKeyPhase2Distribute(
			parames, blind_vec, y_vec, broadcastMsgs1)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		vss_scheme_vec = append(vss_scheme_vec, vss_scheme)
		secret_shares_vec = append(secret_shares_vec, secret_shares)
		index_vec = append(index_vec, index)
	}
	/*
	 */
	vss0 = vss_scheme_vec[0]
	//log.Trace(fmt.Sprintf("secret_shares_vec=%s", utils.StringInterface(secret_shares_vec, 5)))
	var partyShares [][]*big.Int //重新组合的secret shares
	for i := 0; i < n; i++ {
		var s []*big.Int
		for j := 0; j < n; j++ {
			s = append(s, secret_shares_vec[j][i])
		}
		partyShares = append(partyShares, s)
	}
	//log.Trace(fmt.Sprintf("partyShares=%s", utils.StringInterface(partyShares, 5)))

	var dlog_proof_vec []*proofs.DLogProof
	for i := 0; i < n; i++ {
		shared_keys, dlog_proof, err := party_keys_vec[i].phase2_verify_vss_construct_keypair_phase3_pok_dlog(parames,
			y_vec,
			partyShares[i], //所有其他人给第i个公证人的私钥片秘密信息,
			vss_scheme_vec,
			index_vec[i]+1,
		)
		log.Trace(fmt.Sprintf("shared_keys[%d]=%s", i, utils.StringInterface(shared_keys, 3)))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		shared_keys_vec = append(shared_keys_vec, shared_keys)
		dlog_proof_vec = append(dlog_proof_vec, dlog_proof)

	}
	//todo 验证每一个公证人分享出来的组合私钥片信息的有效性.
	if !VerifyDlogProofs(parames, dlog_proof_vec) {
		return nil, nil, nil, nil, nil, errors.New("bad dlog proof")
	}
	//t = n - 1
	var xi []*big.Int
	for i := 0; i < t+1; i++ {
		xi = append(xi, shared_keys_vec[i].xi)
	}
	//todo 这个生成的到底是什么呢? 大家最终分享算出来的私钥是多少?
	x := vss0.Reconstruct(index_vec[0:t+1], xi)
	sum := big.NewInt(0)
	for i := 0; i < len(party_keys_vec); i++ {
		secret_sharing.ModAdd(sum, party_keys_vec[i].ui)
	}
	if x.Cmp(sum) != 0 {
		log.Trace(fmt.Sprintf("x=%s", x.Text(16)))
		log.Trace(fmt.Sprintf("sum=%s", sum.Text(16)))
		return nil, nil, nil, nil, nil, errors.New("sum not equal")
	}
	for i := 0; i < len(dlog_proof_vec); i++ {
		pk_vec = append(pk_vec, dlog_proof_vec[i].PK)
	}
	log.Trace(fmt.Sprintf("part_keys_vec=%s", utils.StringInterface(party_keys_vec, 5)))
	log.Trace(fmt.Sprintf("shared_keys_vec=%s", utils.StringInterface(shared_keys_vec, 7)))
	log.Trace(fmt.Sprintf("pk_vec=%s", utils.StringInterface(pk_vec, 5)))
	log.Trace(fmt.Sprintf("total_pubkey=%s", utils.StringInterface(total_pubkey, 3)))
	log.Trace(fmt.Sprintf("vss0=%s ", utils.StringInterface(vss_scheme_vec[0], 5)))
	return
}
func strDecimal2bigint(s string) *big.Int {
	p, b := new(big.Int).SetString(s, 10)
	if !b {
		panic("conver err")
	}
	return p
}
func strHex2bigint(s string) *big.Int {
	p, b := new(big.Int).SetString(s, 16)
	if !b {
		panic("conver err")
	}
	return p
}
func genKey() (party_keys_vec []*Keys, shared_keys_vec []*SharedKeys, pk_vec []*secret_sharing.GE, total_pubkey *secret_sharing.GE, vss0 *secret_sharing.VerifiableSS) {
	/// 生成五个key
	//1
	k := &Keys{}
	k.ui, _ = new(big.Int).SetString("2f7a3834a7201c47ade1ae354fc0985cee98993dd97b8878d3214745c8fb1a6c", 16)
	yix, yiy := secret_sharing.S.ScalarBaseMult(k.ui.Bytes())
	k.yi = secret_sharing.NewGE(yix, yiy)
	k.dk = proofs.NewPrivateKey(strDecimal2bigint("174626979594475523032569712051907505041181833285378838591284337457725937492907331545234992702280575440921450781215012657867162162164196476088592705047444061953994036726743468332452511219663669929026015130989013800614722711216296873481909944767319296549518785315515579832005867984554172686339215822285029447891"),
		strDecimal2bigint("108035879063423595104087929563149128307930640908148938476711978282186525682938278418063402515933554415508335609937453589085143422419009325361547492990217814356487105779685630145924004139457153217372064228596176356704745164088680225261804503497247935156361239329388946615864762024681829821906511457730628582931"))
	k.partyIndex = 0
	party_keys_vec = append(party_keys_vec, k)
	//2
	k = &Keys{}
	k.ui, _ = new(big.Int).SetString("933563af1c80f436b83eb804d32ae9d0418199c33e218c05c568b1368339a414", 16)
	yix, yiy = secret_sharing.S.ScalarBaseMult(k.ui.Bytes())
	k.yi = secret_sharing.NewGE(yix, yiy)
	k.dk = proofs.NewPrivateKey(strDecimal2bigint("137314806534598616944952852517839799533455683574235631534517255077794708067991431806096814944105737849008867118310891545573053188630998627614573194011961155880719389391710266369793037243810503249918302626949006572319232192860875159997653297862190828959819280572472421751362584260445691160122156900930835007523"),
		strDecimal2bigint("123358272148985205999774944035503275577064376854125703439598938603855959653089103154934935763160161952271697225489056935000571498908312898165280564700733163982956323210270464058063270606941222768721198753166612281749245489155155687877875570681854748862047925198177610060672677085702818644377766345356661649067"))
	k.partyIndex = 1
	party_keys_vec = append(party_keys_vec, k)

	//3
	k = &Keys{}
	k.ui, _ = new(big.Int).SetString("f2fa889c91a4b61ed432eb2acbbbfed4ef9d034aa0d117e2e31074182f7c7d8a", 16)
	yix, yiy = secret_sharing.S.ScalarBaseMult(k.ui.Bytes())
	k.yi = secret_sharing.NewGE(yix, yiy)
	k.dk = proofs.NewPrivateKey(strDecimal2bigint("140375840129417995998526455809425553511706115404720755447623908414290083155898800218778979903054242423242227457275898271537513655788993336195126177428976881468899574581752489957422643483976238321233902551608517155027975783858763157204750142081970843229034130187449284663873142120612748795686467518517136044047"),
		strDecimal2bigint("97258321397525961583792902437942540843615609978463944184073716595572335138267916475069643247708151427411824531556400432093643209032738413984076181365679830377014017706882804388985434564407438671890077549471161822747931776864139519341503652583493334329903526135575652504383611826793037359123858984270891113979"))
	k.partyIndex = 2
	party_keys_vec = append(party_keys_vec, k)

	//4
	k = &Keys{}
	k.ui, _ = new(big.Int).SetString("ae13e7399000421632775816fa5952e606a0ccfd885b0ab18e1787027ca3a3da", 16)
	yix, yiy = secret_sharing.S.ScalarBaseMult(k.ui.Bytes())
	k.yi = secret_sharing.NewGE(yix, yiy)
	k.dk = proofs.NewPrivateKey(strDecimal2bigint("106951107055568857391206526668245186700747197539872937484594530382914725649458536356698188158102787684118439693848146606409628885559328040954890433462366001594223426034205370998262817462949930089540525386530775691806093546215535247340069444390017845739327534818702124122801387388905353562907010155192638607911"),
		strDecimal2bigint("179320191636323321828958508974717690011525651655961819758333239874395031658050861376639014809749453872148850158665293501145726226846462972529242386776033512585544681655596036183763951846966338985252591074167561998353360175748852597427132389197137555949017917134235131460397986744048480015151447614603286908123"))
	k.partyIndex = 3
	party_keys_vec = append(party_keys_vec, k)

	k = &Keys{}
	k.ui, _ = new(big.Int).SetString("57ade70e8b147b8ebbfc25bc7df7b1a56647b49fe939a0f4cc4387f2948ac27d", 16)
	yix, yiy = secret_sharing.S.ScalarBaseMult(k.ui.Bytes())
	k.yi = secret_sharing.NewGE(yix, yiy)
	k.dk = proofs.NewPrivateKey(strDecimal2bigint("141710820404850993438757934057529745768390559317356855543355692671300904185640559637835614666982069037753593833227725044984782018577779689359696130304878425950070946161117312253251326118693702679659563818593899644992153588967459157182857679163723498277649481119457134914888353955921512318200703514290112438821"),
		strDecimal2bigint("118077638327331474411308111476524100720994573518343427837790570787912447359202662499114804440635099778060841625195682216806037778562841038554150808079996912349484199868871843044383286817127210940102969159242462282915820904787367373180349696610138098906506577774708436485769623883608662211826551670548825284539"))
	k.partyIndex = 4
	party_keys_vec = append(party_keys_vec, k)

	//五个sharedKeys
	var sk *SharedKeys
	yix, yiy = secret_sharing.Strtoxy("23946e1ef0bcf8a1b926d445be4f0054bb7f20677d1669a9f31d45e4027a9ae8f5fb9a8c91a2f580e93fa053d2d25310269ded4be2e0516e217382d6822c1ab0")
	sk = &SharedKeys{}
	sk.xi = strHex2bigint("bb6bf2c8705a844228c6cf3866f885901741fe1bcb7197905650be6fec731fee")
	sk.y = secret_sharing.NewGE(yix, yiy)
	shared_keys_vec = append(shared_keys_vec, sk)

	sk = &SharedKeys{}
	sk.xi = strHex2bigint("bb6bf2c8705a844228c6cf3866f885901741fe1bcb7197905650be6fec732011")
	sk.y = secret_sharing.NewGE(yix, yiy)
	shared_keys_vec = append(shared_keys_vec, sk)

	sk = &SharedKeys{}
	sk.xi = strHex2bigint("bb6bf2c8705a844228c6cf3866f885901741fe1bcb7197905650be6fec732048")
	sk.y = secret_sharing.NewGE(yix, yiy)
	shared_keys_vec = append(shared_keys_vec, sk)

	sk = &SharedKeys{}
	sk.xi = strHex2bigint("bb6bf2c8705a844228c6cf3866f885901741fe1bcb7197905650be6fec732093")
	sk.y = secret_sharing.NewGE(yix, yiy)
	shared_keys_vec = append(shared_keys_vec, sk)

	sk = &SharedKeys{}
	sk.xi = strHex2bigint("bb6bf2c8705a844228c6cf3866f885901741fe1bcb7197905650be6fec7320f2")
	sk.y = secret_sharing.NewGE(yix, yiy)
	shared_keys_vec = append(shared_keys_vec, sk)

	yix, yiy = secret_sharing.Strtoxy("23946e1ef0bcf8a1b926d445be4f0054bb7f20677d1669a9f31d45e4027a9ae8f5fb9a8c91a2f580e93fa053d2d25310269ded4be2e0516e217382d6822c1ab0")
	total_pubkey = secret_sharing.NewGE(yix, yiy)

	vss0 = &secret_sharing.VerifiableSS{
		Parameters:  secret_sharing.ShamirSecretSharing{2, 5},
		Commitments: make([]*secret_sharing.GE, 3),
	}
	yix, yiy = secret_sharing.Strtoxy("4c25ebabf0d4f79c6373520d66c05fa00ed5d316279419d6f3550e356b1ca17dd3c591c492f9c42cbc9532e4e39844f1fbe472b842666390c45ee0d5456aa2d3")
	vss0.Commitments[0] = secret_sharing.NewGE(yix, yiy)
	yix, yiy = secret_sharing.Strtoxy("9817f8165b81f259d928ce2ddbfc9b02070b87ce9562a055acbbdcf97e66be79b8d410fb8fd0479c195485a648b417fda808110efcfba45d65c4a32677da3a48")
	vss0.Commitments[1] = secret_sharing.NewGE(yix, yiy)
	yix, yiy = secret_sharing.Strtoxy("e59e705cb909acaba73cef8c4b8e775cd87cc0956e4045306d7ded41947f04c62ae5cf50a9316423e1d066326532f6f7eeea6c461984c5a339c33da6fe68e11a")
	vss0.Commitments[2] = secret_sharing.NewGE(yix, yiy)
	yix.String()
	return
}
func Sign(t, n, ttag int, s []int) error {
	var err error
	//party_keys_vec, shared_keys_vec, _, y, vss_scheme, err := KeyGenTN(t, n)
	//if err != nil {
	//	return err
	//}
	party_keys_vec, shared_keys_vec, _, y, vss_scheme := genKey()
	log.Trace(fmt.Sprintf("party_keys_vec=%s", utils.StringInterface(party_keys_vec, 5)))
	log.Trace(fmt.Sprintf("shared_keys_vec=%s", utils.StringInterface(shared_keys_vec, 5)))
	log.Trace(fmt.Sprintf("y=%s", secret_sharing.Xytostr(y.X, y.Y)))
	log.Trace(fmt.Sprintf("vss=%s", utils.StringInterface(vss_scheme, 5)))
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
	var tx, ty *big.Int
	sign_keys_vec[0].ki = strHex2bigint("b33478cfe5717369cda7ad324f433f94e94dd351236996d54974b9c476839e8a")
	sign_keys_vec[0].gammaI = strHex2bigint("1aa79e0dd371048b867ca12d5ecf9d120795146570c20a69a877db8f64e5bb51")
	tx, ty = secret_sharing.Strtoxy("ea177ffb3c92346ef229a50d8ed53459f4857f3ccddddc5ff12ad8da55d29300313a58032723460dce5e8881fd97bceff3d04a429814fad8dd62974f0ae0be06")
	sign_keys_vec[0].gGammaI = secret_sharing.NewGE(tx, ty)

	sign_keys_vec[1].ki = strHex2bigint("f7be401f0800ca965b446d5bd6966bf7bc94bec8d5f43d2e2b4f5dcbe5665b6b")
	sign_keys_vec[1].gammaI = strHex2bigint("77fb64a960cfa225c3035bcfe2907bd5fedc0b632b9b346cfa859580cc83eede")
	tx, ty = secret_sharing.Strtoxy("da6bea1a140421992daa74230b18896a7a03128d1efce22e2e7bf168daa0ba27290c1e3c031e6c17139f2a465f42b483fdf5f8f79a0950a097be0abef15eb830")
	sign_keys_vec[1].gGammaI = secret_sharing.NewGE(tx, ty)

	sign_keys_vec[2].ki = strHex2bigint("9e5118fc0fc1fecad68a727d723eab8608a8420d95b7af92456396423adaadb5")
	sign_keys_vec[2].gammaI = strHex2bigint("05985e6c2aa6859b07cab846e339e2edf6cb44d1baaf41352e04738b0df4db1d")
	tx, ty = secret_sharing.Strtoxy("c2b6f70f6de4148cf2be87772178d278c4d9c1ed5e5f998fc2990aa769cf06d03b2e7232a5b2b57556e9553dcfd07b0055f1a9d146906036357f780ad273da78")
	sign_keys_vec[2].gGammaI = secret_sharing.NewGE(tx, ty)

	sign_keys_vec[3].ki = strHex2bigint("1d6bb049eb7cc4885eec57bc36f2cc8210ba7cf97ee65a735d977bb0591197b1")
	sign_keys_vec[3].gammaI = strHex2bigint("bed1667cb4ce4a3379248ca21dc8f21db6004dc19804dbde03b6c90b419bd7f3")
	tx, ty = secret_sharing.Strtoxy("ed94a04e2594257b2159c30dd4a7ed168ba3d63221f7a877243e3ef0c047d5136a14b4f74583a4ec62c0f39d628f31a5d506d0503a492c1ffad17187f86a6e26")
	sign_keys_vec[3].gGammaI = secret_sharing.NewGE(tx, ty)

	log.Trace(fmt.Sprintf("sign_keys_vec=%s", utils.StringInterface(sign_keys_vec, 5)))
	// each party computes [Ci,Di] = com(g^gamma_i) and broadcast the commitmentsvar
	var bc1_vec []*SignBroadcastPhase1
	var blind_vec1 []*big.Int
	for i := 0; i < ttag; i++ {
		var b *big.Int
		switch i {
		case 0:
			b = strDecimal2bigint("80158307830119780958043459566864686385976285004168522580396243240100053191930")
		case 1:
			b = strDecimal2bigint("44044565582300379798999175574250065737439086120865718102731524008411910113719")
		case 2:
			b = strDecimal2bigint("20109803869838010947049125485151376760131658485847720191949103864821272889816")
		case 3:
			b = strDecimal2bigint("12015280755751774205851693677859634032372492471049750671687268029743216584989")
		}
		com, blind := sign_keys_vec[i].testPhase1Broadcast(b)
		bc1_vec = append(bc1_vec, com)
		blind_vec1 = append(blind_vec1, blind)
	}
	log.Trace(fmt.Sprintf("bc1_vec=%s", utils.StringInterface(bc1_vec, 5)))
	log.Trace(fmt.Sprintf("blind_vec1=%s", utils.StringInterface(blind_vec1, 5)))
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
	log.Trace(fmt.Sprintf("m_a_vec=%s", utils.StringInterface(m_a_vec, 5)))
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
	log.Trace(fmt.Sprintf("m_b_gamma_vec_all=%s", utils.StringInterface(m_b_gamma_vec_all, 5)))
	log.Trace(fmt.Sprintf("beta_vec_all=%s", utils.StringInterface(beta_vec_all, 5)))
	log.Trace(fmt.Sprintf("m_b_w_vec_all=%s", utils.StringInterface(m_b_w_vec_all, 5)))
	log.Trace(fmt.Sprintf("ni_vec_all=%s", utils.StringInterface(ni_vec_all, 5)))

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
		m_b_w_vec_i := m_b_w_vec_all[i]
		for j := 0; j < ttag-1; j++ {
			ind := j
			if j >= i {
				ind = j + 1
			}

			mb := m_b_gamma_vec_i[j]
			log.Trace(fmt.Sprintf("mb=%s", utils.StringInterface(mb, 5)))
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
	log.Trace(fmt.Sprintf("alpha_vec_all=%s", utils.StringInterface(alpha_vec_all, 5)))
	log.Trace(fmt.Sprintf("miu_vec_all=%s", utils.StringInterface(miu_vec_all, 5)))

	var delta_vec []*big.Int
	var sigma_vec []*big.Int

	for i := 0; i < ttag; i++ {
		delta := sign_keys_vec[i].phase2DeltaI(alpha_vec_all[i], beta_vec_all[i])
		sigma := sign_keys_vec[i].phase2SigmaI(miu_vec_all[i], ni_vec_all[i])
		delta_vec = append(delta_vec, delta)
		sigma_vec = append(sigma_vec, sigma)
	}
	log.Trace(fmt.Sprintf("delta_vec=%s", utils.StringInterface(delta_vec, 5)))
	log.Trace(fmt.Sprintf("sigma_vec=%s", sigma_vec))

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
	log.Trace(fmt.Sprintf("R_vec=%s", utils.StringInterface(r_vec, 5)))

	message := []byte{79, 77, 69, 82}
	messageHash := utils.ShaSecret(message)
	messageBN := new(big.Int).SetBytes(messageHash[:])
	var local_sig_vec []*LocalSignature
	// each party computes s_i but don't send it yet. we start with phase5

	for i := 0; i < ttag; i++ {
		localSig := phase5LocalSignature(sign_keys_vec[i].ki, messageBN, r_vec[i], sigma_vec[i], y)
		local_sig_vec = append(local_sig_vec, localSig)
	}
	log.Trace(fmt.Sprintf("local_sig_vec=%s", utils.StringInterface(local_sig_vec, 5)))
	var phase5_com_vec []*Phase5Com1
	var phase5a_decom_vec []*Phase5ADecom1
	var helgamal_proof_vec []*proofs.HomoELGamalProof
	// we notice that the proof for V= R^sg^l, B = A^l is a general form of homomorphic elgamal.
	for i := 0; i < ttag; i++ {
		var blindFactor *big.Int
		switch i {
		case 0:
			blindFactor = strDecimal2bigint("23132629008796396900966727937771914610780869746500214102950787078128504078915")
		case 1:
			blindFactor = strDecimal2bigint("48208059360612339533547262280451089093124132645349488651880489948786504352315")
		case 2:
			blindFactor = strDecimal2bigint("8955404397181861341094241365273270692152750178515180707982760591669003061542")
		case 3:
			blindFactor = strDecimal2bigint("72125768621773903647254546210426739597001283324644430540014241088136847934613")

		}
		phase5Com, phase5ADecom, helgamalProof := local_sig_vec[i].testphase5aBroadcast5bZkproof(blindFactor)
		phase5_com_vec = append(phase5_com_vec, phase5Com)
		phase5a_decom_vec = append(phase5a_decom_vec, phase5ADecom)
		helgamal_proof_vec = append(helgamal_proof_vec, helgamalProof)
	}
	log.Trace(fmt.Sprintf("phase5_com_vec=%s", utils.StringInterface(phase5_com_vec, 5)))
	log.Trace(fmt.Sprintf("phase_5a_decom_vec=%s", utils.StringInterface(phase5a_decom_vec, 5)))
	log.Trace(fmt.Sprintf("helgamal_proof_vec=%s", utils.StringInterface(helgamal_proof_vec, 5)))

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
		var blindFactor *big.Int
		switch i {
		case 0:
			blindFactor = strDecimal2bigint("102693997887396385491204830191083969170786846118205855244598142170969005560227")
		case 1:
			blindFactor = strDecimal2bigint("93325289958969391723610594203726837204198168911978574962643114563372979001239")

		case 2:
			blindFactor = strDecimal2bigint("14837766750661644843410459725346758216136419724963965342273388950253109667574")
		case 3:
			blindFactor = strDecimal2bigint("36010640766639678313227722844811757280259207782856913930292101889964106925955")

		}
		phase5com2, phase5ddecom2, err := local_sig_vec[i].testphase5c(
			decom_i, com_i,
			elgamla_i, phase5a_decom_vec[i].vi, r_vec[0], blindFactor,
		)
		if err != nil {
			return fmt.Errorf("error phase5 err %s", err)
		}
		phase5_com2_vec = append(phase5_com2_vec, phase5com2)
		phase_5d_decom2_vec = append(phase_5d_decom2_vec, phase5ddecom2)
	}
	log.Trace(fmt.Sprintf("phase5_com2_vec=%s", utils.StringInterface(phase5_com2_vec, 5)))
	log.Trace(fmt.Sprintf("phase_5d_decom2_vec=%s", utils.StringInterface(phase_5d_decom2_vec, 5)))

	// assuming phase5 checks passes each party sends s_i and compute sum_i{s_i}
	var s_vec []*big.Int
	for i := 0; i < ttag; i++ {
		si, err := local_sig_vec[i].phase5d(phase_5d_decom2_vec, phase5_com2_vec, phase5a_decom_vec)
		if err != nil {
			return fmt.Errorf("bad com 5d %s", err)
		}
		s_vec = append(s_vec, si)
	}
	log.Trace(fmt.Sprintf("s_vec=%s", utils.StringInterface(s_vec, 5)))
	// here we compute the signature only of party i=0 to demonstrate correctness.
	_, _, err = local_sig_vec[0].outputSignature(s_vec[1:])
	if err != nil {
		return fmt.Errorf("sigature verify  err %s", err)
	}
	return nil
}

func Sign2(t, n, ttag int, s []int) error {
	var err error
	party_keys_vec, shared_keys_vec, _, y, vss_scheme, err := KeyGenTN(t, n)
	if err != nil {
		return err
	}
	log.Trace(fmt.Sprintf("party_keys_vec=%s", utils.StringInterface(party_keys_vec, 5)))
	log.Trace(fmt.Sprintf("shared_keys_vec=%s", utils.StringInterface(shared_keys_vec, 5)))
	log.Trace(fmt.Sprintf("y=%s", secret_sharing.Xytostr(y.X, y.Y)))
	log.Trace(fmt.Sprintf("vss=%s", utils.StringInterface(vss_scheme, 5)))
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

	log.Trace(fmt.Sprintf("sign_keys_vec=%s", utils.StringInterface(sign_keys_vec, 5)))
	// each party computes [Ci,Di] = com(g^gamma_i) and broadcast the commitmentsvar
	var bc1_vec []*SignBroadcastPhase1
	var blind_vec1 []*big.Int
	for i := 0; i < ttag; i++ {
		com, blind := sign_keys_vec[i].phase1Broadcast()
		bc1_vec = append(bc1_vec, com)
		blind_vec1 = append(blind_vec1, blind)
	}
	log.Trace(fmt.Sprintf("bc1_vec=%s", utils.StringInterface(bc1_vec, 5)))
	log.Trace(fmt.Sprintf("blind_vec1=%s", utils.StringInterface(blind_vec1, 5)))
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
	log.Trace(fmt.Sprintf("m_a_vec=%s", utils.StringInterface(m_a_vec, 5)))
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
	log.Trace(fmt.Sprintf("m_b_gamma_vec_all=%s", utils.StringInterface(m_b_gamma_vec_all, 5)))
	log.Trace(fmt.Sprintf("beta_vec_all=%s", utils.StringInterface(beta_vec_all, 5)))
	log.Trace(fmt.Sprintf("m_b_w_vec_all=%s", utils.StringInterface(m_b_w_vec_all, 5)))
	log.Trace(fmt.Sprintf("ni_vec_all=%s", utils.StringInterface(ni_vec_all, 5)))

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
		m_b_w_vec_i := m_b_w_vec_all[i]
		for j := 0; j < ttag-1; j++ {
			ind := j
			if j >= i {
				ind = j + 1
			}

			mb := m_b_gamma_vec_i[j]
			log.Trace(fmt.Sprintf("mb=%s", utils.StringInterface(mb, 5)))
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
	log.Trace(fmt.Sprintf("alpha_vec_all=%s", utils.StringInterface(alpha_vec_all, 5)))
	log.Trace(fmt.Sprintf("miu_vec_all=%s", utils.StringInterface(miu_vec_all, 5)))

	var delta_vec []*big.Int
	var sigma_vec []*big.Int

	for i := 0; i < ttag; i++ {
		delta := sign_keys_vec[i].phase2DeltaI(alpha_vec_all[i], beta_vec_all[i])
		sigma := sign_keys_vec[i].phase2SigmaI(miu_vec_all[i], ni_vec_all[i])
		delta_vec = append(delta_vec, delta)
		sigma_vec = append(sigma_vec, sigma)
	}
	log.Trace(fmt.Sprintf("delta_vec=%s", utils.StringInterface(delta_vec, 5)))
	log.Trace(fmt.Sprintf("sigma_vec=%s", sigma_vec))

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
	log.Trace(fmt.Sprintf("R_vec=%s", utils.StringInterface(r_vec, 5)))

	message := []byte{79, 77, 69, 82}
	messageHash := utils.ShaSecret(message)
	messageBN := new(big.Int).SetBytes(messageHash[:])
	var local_sig_vec []*LocalSignature
	// each party computes s_i but don't send it yet. we start with phase5

	for i := 0; i < ttag; i++ {
		localSig := phase5LocalSignature(sign_keys_vec[i].ki, messageBN, r_vec[i], sigma_vec[i], y)
		local_sig_vec = append(local_sig_vec, localSig)
	}
	log.Trace(fmt.Sprintf("local_sig_vec=%s", utils.StringInterface(local_sig_vec, 5)))
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
	log.Trace(fmt.Sprintf("phase5_com_vec=%s", utils.StringInterface(phase5_com_vec, 5)))
	log.Trace(fmt.Sprintf("phase_5a_decom_vec=%s", utils.StringInterface(phase5a_decom_vec, 5)))
	log.Trace(fmt.Sprintf("helgamal_proof_vec=%s", utils.StringInterface(helgamal_proof_vec, 5)))

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
			decom_i, com_i,
			elgamla_i, phase5a_decom_vec[i].vi, r_vec[0],
		)
		if err != nil {
			return fmt.Errorf("error phase5 err %s", err)
		}
		phase5_com2_vec = append(phase5_com2_vec, phase5com2)
		phase_5d_decom2_vec = append(phase_5d_decom2_vec, phase5ddecom2)
	}
	log.Trace(fmt.Sprintf("phase5_com2_vec=%s", utils.StringInterface(phase5_com2_vec, 5)))
	log.Trace(fmt.Sprintf("phase_5d_decom2_vec=%s", utils.StringInterface(phase_5d_decom2_vec, 5)))

	// assuming phase5 checks passes each party sends s_i and compute sum_i{s_i}
	var s_vec []*big.Int
	for i := 0; i < ttag; i++ {
		si, err := local_sig_vec[i].phase5d(phase_5d_decom2_vec, phase5_com2_vec, phase5a_decom_vec)
		if err != nil {
			return fmt.Errorf("bad com 5d %s", err)
		}
		s_vec = append(s_vec, si)
	}
	log.Trace(fmt.Sprintf("s_vec=%s", utils.StringInterface(s_vec, 5)))
	// here we compute the signature only of party i=0 to demonstrate correctness.
	_, _, err = local_sig_vec[0].outputSignature(s_vec[1:])
	if err != nil {
		return fmt.Errorf("sigature verify  err %s", err)
	}
	return nil
}
