package dcrm

import (
	"os"
	"testing"

	"crypto/rand"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/models"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"

	"encoding/json"

	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, utils.MyStreamHandler(os.Stderr)))
}

var shareArg *NotaryShareArg

func init() {
	shareArg = &NotaryShareArg{

		MasterPK: commitments.GenerateNMMasterPublicKey(),
	}
	shareArg.PaillierPrivateKey, _ = zkp.GenerateKey(rand.Reader, 1023)
	shareArg.ZkPublicParams = zkp.GenerateParams(G, 256, 512, &shareArg.PaillierPrivateKey.PublicKey)
}
func newTestNotaryShareArg(name string) *NotaryShareArg {
	n := &NotaryShareArg{
		MasterPK:           shareArg.MasterPK,
		PaillierPrivateKey: shareArg.PaillierPrivateKey,
		ZkPublicParams:     shareArg.ZkPublicParams,
	}
	n.Name = name
	return n
}
func newOneNotarySerice() *NotaryService {
	return newNotaryService("n1")
}
func newNotaryService(name string) *NotaryService {
	k, _ := crypto.GenerateKey()
	return &NotaryService{
		NotaryShareArg: newTestNotaryShareArg(name),
		Notaries: map[string]*NotatoryInfo{
			name: {
				Name: name,
				Key:  &k.PublicKey,
			},
		},
	}
}
func newThreeNotaryService() (ns1, ns2, ns3 *NotaryService) {
	k, _ := crypto.GenerateKey()
	n1 := "n1"
	n2 := "n2"
	n3 := "n3"
	ns1 = &NotaryService{
		NotaryShareArg: newTestNotaryShareArg(n1),
		Notaries: map[string]*NotatoryInfo{
			n1: {
				Name: n1,
				Key:  &k.PublicKey,
			},
			n2: {
				Name: n2,
				Key:  &k.PublicKey,
			},
			n3: {
				Name: n3,
				Key:  &k.PublicKey,
			},
		},
	}
	ns2 = &NotaryService{
		NotaryShareArg: newTestNotaryShareArg(n2),
		Notaries:       ns1.Notaries,
	}
	ns3 = &NotaryService{
		NotaryShareArg: newTestNotaryShareArg(n3),
		Notaries:       ns1.Notaries,
	}
	return
}

func newLockedIn() *models.PrivateKeyInfo {
	db := models.SetupTestDB()
	ns1, ns2, ns3 := newThreeNotaryService()
	l1 := &LockedIn{
		db:  db,
		srv: ns1,
		Key: utils.NewRandomHash(),
	}
	l2 := &LockedIn{
		srv: ns2,
		Key: l1.Key,
	}
	l3 := &LockedIn{
		srv: ns3,
		Key: l1.Key,
	}
	_, err := l1.LockinKeyGenerate()
	if err != nil {
		panic(err)
	}
	pi2 := l2.doLockinKeyGenerate()
	pi3 := l3.doLockinKeyGenerate()
	finish, err := l1.AddNewZkpFromOtherNotary(pi2.Zkps[0])
	if err != nil {
		panic(err)
	}

	finish, err = l1.AddNewZkpFromOtherNotary(pi3.Zkps[0])
	if err != nil {
		panic(err)
	}
	if !finish {
		panic("should finish")
	}
	pi, err := l1.db.LoadPrivatedKeyInfo(l1.Key)
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("总加密私钥=%s\n,总公钥x=%s\npbubkeyy=%s",
		pi.EncryptedPrivateKey.String(), pi.PublicKeyX.String(),
		pi.PublicKeyY.String()))
	return pi
}

func TestLockedIn_LockinKeyGenerate(t *testing.T) {
	db := models.SetupTestDB()
	l := &LockedIn{
		db:  db,
		srv: newOneNotarySerice(),
		Key: utils.NewRandomHash(),
	}
	z, err := l.LockinKeyGenerate()
	if err != nil {
		t.Error(err)
	}
	t.Logf("zkp=%s", utils.StringInterface(z, 7))
	z2, err := db.LoadPrivatedKeyInfo(l.Key)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("z2=%s", utils.StringInterface(z2, 5))
	assert.EqualValues(t, len(z2.Zkps), 1)
}

func TestLockedIn_AddNewZkpFromOtherNotary(t *testing.T) {
	db := models.SetupTestDB()
	ns1, ns2, ns3 := newThreeNotaryService()
	l1 := &LockedIn{
		db:  db,
		srv: ns1,
		Key: utils.NewRandomHash(),
	}
	l2 := &LockedIn{
		srv: ns2,
		Key: l1.Key,
	}
	l3 := &LockedIn{
		srv: ns3,
		Key: l1.Key,
	}
	z, err := l1.LockinKeyGenerate()
	if err != nil {
		t.Error(err)
	}
	t.Logf("zkp=%s", utils.StringInterface(z, 7))
	pi2 := l2.doLockinKeyGenerate()
	pi3 := l3.doLockinKeyGenerate()
	finish, err := l1.AddNewZkpFromOtherNotary(pi2.Zkps[0])
	assert.EqualValues(t, finish, false)
	assert.EqualValues(t, err, nil)
	finish, err = l1.AddNewZkpFromOtherNotary(pi2.Zkps[0])
	assert.EqualValues(t, finish, false)
	assert.EqualValues(t, err != nil, true)
	finish, err = l1.AddNewZkpFromOtherNotary(pi3.Zkps[0])
	assert.EqualValues(t, finish, true)
	assert.EqualValues(t, err != nil, false)
	pi3.Zkps[0].NotaryName = "randxxx"
	finish, err = l1.AddNewZkpFromOtherNotary(pi3.Zkps[0])
	//已经验证完毕,不能再修改
	assert.EqualValues(t, err != nil, true)
	pi, err := db.LoadPrivatedKeyInfo(l1.Key)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, len(pi.Zkps), 3)
	assert.EqualValues(t, pi.EncryptedPrivateKey != nil, true)
	assert.EqualValues(t, pi.PublicKeyX != nil, true)
	assert.EqualValues(t, pi.PublicKeyY != nil, true)
}

func TestMarshalArg(t *testing.T) {
	n := newTestNotaryShareArg("n1")
	data, err := json.Marshal(n)
	if err != nil {
		t.Error(err)
		return
	}
	n2 := new(NotaryShareArg)
	json.Unmarshal(data, n2)
	t.Logf("n=%s", utils.StringInterface(n, 7))
	t.Logf("n2=%s", utils.StringInterface(n2, 7))
}

func TestDB_NewPrivateKeyInfo(t *testing.T) {
	pi := newLockedIn()
	t.Logf("x=%s,y=%s", pi.PublicKeyX, pi.PublicKeyY)
	key, _ := crypto.GenerateKey()
	pub := key.PublicKey
	pub.X = pi.PublicKeyX
	pub.Y = pi.PublicKeyY
	addr := crypto.PubkeyToAddress(pub)
	t.Logf("addr=%s", addr.String())
}

func TestLockedIn_AddNewZkpFromOtherNotary2(t *testing.T) {
	ns1, ns2, ns3 := newThreeNotaryService()
	l1 := &LockedIn{
		db:  models.SetupTestDB2("n1"),
		srv: ns1,
		Key: utils.NewRandomHash(),
	}
	l2 := &LockedIn{
		db:  models.SetupTestDB2("n2"),
		srv: ns2,
		Key: l1.Key,
	}
	l3 := &LockedIn{
		db:  models.SetupTestDB2("n3"),
		srv: ns3,
		Key: l1.Key,
	}
	z, err := l1.LockinKeyGenerate()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("zkp=%s", utils.StringInterface(z, 7))
	z2, err := l2.LockinKeyGenerate()
	if err != nil {
		t.Error(err)
		return
	}
	z3, err := l3.LockinKeyGenerate()
	if err != nil {
		t.Error(err)
		return
	}
	//1 添加2,3进行验证
	finish, err := l1.AddNewZkpFromOtherNotary(z2)
	assert.EqualValues(t, finish, false)
	assert.EqualValues(t, err, nil)
	finish, err = l1.AddNewZkpFromOtherNotary(z2)
	assert.EqualValues(t, finish, false)
	assert.EqualValues(t, err != nil, true)
	finish, err = l1.AddNewZkpFromOtherNotary(z3)
	assert.EqualValues(t, finish, true)
	assert.EqualValues(t, err, nil)

	//2 添加1,3
	finish, err = l2.AddNewZkpFromOtherNotary(z)
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	finish, err = l2.AddNewZkpFromOtherNotary(z3)
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//3 添加1,2
	finish, err = l3.AddNewZkpFromOtherNotary(z)
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	finish, err = l3.AddNewZkpFromOtherNotary(z2)
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//已经验证完毕,不能再修改
	pi1, _ := l1.db.LoadPrivatedKeyInfo(l1.Key)
	pi2, _ := l2.db.LoadPrivatedKeyInfo(l2.Key)
	pi3, _ := l3.db.LoadPrivatedKeyInfo(l3.Key)
	assert.EqualValues(t, pi1.PublicKeyX, pi2.PublicKeyX)
	assert.EqualValues(t, pi2.PublicKeyX, pi3.PublicKeyX)
	assert.EqualValues(t, pi1.PublicKeyY, pi2.PublicKeyY)
	assert.EqualValues(t, pi2.PublicKeyY, pi3.PublicKeyY)
	assert.EqualValues(t, pi1.EncryptedPrivateKey, pi2.EncryptedPrivateKey)
	assert.EqualValues(t, pi2.EncryptedPrivateKey, pi2.EncryptedPrivateKey)
}
