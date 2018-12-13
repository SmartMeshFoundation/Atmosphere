package dcrm2

import (
	"testing"

	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/dcrm/models"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var addr common.Address

func newThreeLockedOut() (l1, l2, l3 *LockedOut) {
	ns1, ns2, ns3 := newThreeNotaryService()
	db1 := models.SetupTestDB2("n1.db")
	db2 := models.SetupTestDB2("n2.db")
	db3 := models.SetupTestDB2("n3.db")
	key := utils.NewRandomHash()
	pi := newLockedIn()
	addr = calcAddr(pi.PublicKeyX, pi.PublicKeyY)
	log.Info(fmt.Sprintf("addr=%s", addr.String()))
	l1 = &LockedOut{
		db:                  db1,
		srv:                 ns1,
		Key:                 key,
		EncryptedPrivateKey: pi.EncryptedPrivateKey,
		Message:             []byte("aaa"),
		pi:                  pi,
	}
	l2 = &LockedOut{
		db:                  db2,
		srv:                 ns2,
		Key:                 key,
		Message:             l1.Message,
		EncryptedPrivateKey: l1.EncryptedPrivateKey,
		pi:                  pi,
	}
	l3 = &LockedOut{
		db:                  db3,
		srv:                 ns3,
		Key:                 key,
		Message:             l1.Message,
		EncryptedPrivateKey: l1.EncryptedPrivateKey,
		pi:                  pi,
	}
	return
}
func TestLockedOut_NewTx(t *testing.T) {
	l1, l2, l3 := newThreeLockedOut()
	//l2.Message = []byte("ccc")
	tx1, err := l1.NewTx()
	if err != nil {
		t.Error(err)
		return
	}
	tx2, err := l2.NewTx()
	if err != nil {
		t.Error(err)
		return
	}
	tx3, err := l3.NewTx()
	if err != nil {
		t.Error(err)
		return
	}
	//1 添加2,3,
	finish, err := l1.AddNewZkpi1FromOtherNotary(tx2.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)

	finish, err = l1.AddNewZkpi1FromOtherNotary(tx3.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//2,添加1,3
	finish, err = l2.AddNewZkpi1FromOtherNotary(tx1.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)

	finish, err = l2.AddNewZkpi1FromOtherNotary(tx3.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//3添加1,2

	finish, err = l3.AddNewZkpi1FromOtherNotary(tx1.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)

	finish, err = l3.AddNewZkpi1FromOtherNotary(tx2.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	tx1, err = l1.StartLockoutStage2()
	assert.EqualValues(t, err, nil)
	tx2, err = l2.StartLockoutStage2()
	assert.EqualValues(t, err, nil)
	tx3, err = l3.StartLockoutStage2()
	assert.EqualValues(t, err, nil)

	//1 添加2,3
	_, finish, err = l1.AddZkpi2(tx2.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	ntx1, finish, err := l1.AddZkpi2(tx3.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//2添加1,3
	_, finish, err = l2.AddZkpi2(tx1.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	ntx2, finish, err := l2.AddZkpi2(tx3.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//3添加1,2
	_, finish, err = l3.AddZkpi2(tx1.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	ntx3, finish, err := l3.AddZkpi2(tx2.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//1校验签名
	sig1 := l1.CalcSignature(ntx1)
	addr1, err := utils.Ecrecover(utils.Sha3(l1.Message), sig1)
	assert.EqualValues(t, err, nil)
	//
	//l1.Message = []byte("bbb")
	//sig1_2 := l1.CalcSignature(nil)
	//addr1_2, err := utils.Ecrecover(utils.Sha3(l1.Message), sig1_2)
	//assert.EqualValues(t, err, nil)
	//assert.EqualValues(t, addr1_2, addr1)
	//2校验签名
	//l2.Message = []byte("ccc")
	sig2 := l2.CalcSignature(ntx2)

	addr2, err := utils.Ecrecover(utils.Sha3(l2.Message), sig2)
	assert.EqualValues(t, err, nil)

	//3校验签名
	//l3.Message = []byte("ddd")
	sig3 := l3.CalcSignature(ntx3)

	addr3, err := utils.Ecrecover(utils.Sha3(l3.Message), sig3)
	assert.EqualValues(t, err, nil)

	assert.EqualValues(t, addr1, addr)
	assert.EqualValues(t, addr2, addr)
	assert.EqualValues(t, addr3, addr)
	log.Info(fmt.Sprintf("addr1=%s,addr2=%s,addr3=%s", addr1.String(), addr2.String(), addr3.String()))
}

func TestLockedOut_NewTx2(t *testing.T) {
	l1, l2, l3 := newThreeLockedOut()
	//l2.Message = []byte("ccc")
	tx1, err := l1.NewTx()
	if err != nil {
		t.Error(err)
		return
	}
	tx2, err := l2.NewTx()
	if err != nil {
		t.Error(err)
		return
	}
	tx3, err := l3.NewTx()
	if err != nil {
		t.Error(err)
		return
	}
	//1 添加2,3,
	finish, err := l1.AddNewZkpi1FromOtherNotary(tx2.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)

	finish, err = l1.AddNewZkpi1FromOtherNotary(tx3.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//2,添加1,3
	finish, err = l2.AddNewZkpi1FromOtherNotary(tx1.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)

	finish, err = l2.AddNewZkpi1FromOtherNotary(tx3.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//3添加1,2

	finish, err = l3.AddNewZkpi1FromOtherNotary(tx1.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)

	finish, err = l3.AddNewZkpi1FromOtherNotary(tx2.Zkpi1s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	tx1, err = l1.StartLockoutStage2()
	assert.EqualValues(t, err, nil)
	tx2, err = l2.StartLockoutStage2()
	assert.EqualValues(t, err, nil)
	tx3, err = l3.StartLockoutStage2()
	assert.EqualValues(t, err, nil)

	//1 添加2,3
	_, finish, err = l1.AddZkpi2(tx2.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	ntx1, finish, err := l1.AddZkpi2(tx3.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//2添加1,3
	_, finish, err = l2.AddZkpi2(tx1.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	ntx2, finish, err := l2.AddZkpi2(tx3.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//3添加1,2
	_, finish, err = l3.AddZkpi2(tx1.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, false)
	ntx3, finish, err := l3.AddZkpi2(tx2.Zkpi2s[0])
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, finish, true)

	//1校验签名
	sig1 := l1.CalcSignature(ntx1)
	addr1, err := utils.Ecrecover(utils.Sha3(l1.Message), sig1)
	assert.EqualValues(t, err, nil)

	//l1.Message = []byte("bbb")
	//sig1_2 := l1.CalcSignature(nil)
	//addr1_2, err := utils.Ecrecover(utils.Sha3(l1.Message), sig1_2)
	//assert.EqualValues(t, err, nil)
	//assert.EqualValues(t, addr1_2, addr1)
	//2校验签名
	l2.Message = []byte("ccc")
	sig2 := l2.CalcSignature(ntx2)

	addr2, err := utils.Ecrecover(utils.Sha3(l2.Message), sig2)
	assert.EqualValues(t, err, nil)

	//3校验签名
	l3.Message = []byte("ddd")
	sig3 := l3.CalcSignature(ntx3)

	addr3, err := utils.Ecrecover(utils.Sha3(l3.Message), sig3)
	assert.EqualValues(t, err, nil)

	assert.EqualValues(t, addr1, addr)
	assert.EqualValues(t, addr2, addr)
	assert.EqualValues(t, addr3, addr)
	log.Info(fmt.Sprintf("addr1=%s,addr2=%s,addr3=%s", addr1.String(), addr2.String(), addr3.String()))
}

func TestLockedOut_NewTx3(t *testing.T) {
	l1, _, _ := newThreeLockedOut()
	//l2.Message = []byte("ccc")
	tx1, err := l1.newTx2()
	if err != nil {
		t.Error(err)
		return
	}
	tx2, err := l1.newTx2()
	if err != nil {
		t.Error(err)
		return
	}
	tx3, err := l1.newTx2()
	if err != nil {
		t.Error(err)
		return
	}

	tx1, err = l1.StartLockoutStage2Test()
	assert.EqualValues(t, err, nil)
	tx2, err = l1.StartLockoutStage2Test()
	tx3, err = l1.StartLockoutStage2Test()

	//1校验签名
	sig1 := l1.CalcSignature(nil)
	addr1, err := utils.Ecrecover(utils.Sha3(l1.Message), sig1)
	assert.EqualValues(t, err, nil)
	assert.EqualValues(t, tx1 != nil, true)
	assert.EqualValues(t, tx2 != nil, true)
	assert.EqualValues(t, tx3 != nil, true)
	log.Info(fmt.Sprintf("addr1=%s", addr1.String()))
}

func TestLockedOut_NewTx4(t *testing.T) {
	cnt := 0
	for i := 1; i < 10; i++ {
		l1, l2, l3 := newThreeLockedOut()
		//l2.Message = []byte("ccc")
		tx1, err := l1.NewTx()
		if err != nil {
			t.Error(err)
			return
		}
		tx2, err := l2.NewTx()
		if err != nil {
			t.Error(err)
			return
		}
		tx3, err := l3.NewTx()
		if err != nil {
			t.Error(err)
			return
		}
		//1 添加2,3,
		finish, err := l1.AddNewZkpi1FromOtherNotary(tx2.Zkpi1s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, false)

		finish, err = l1.AddNewZkpi1FromOtherNotary(tx3.Zkpi1s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, true)

		//2,添加1,3
		finish, err = l2.AddNewZkpi1FromOtherNotary(tx1.Zkpi1s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, false)

		finish, err = l2.AddNewZkpi1FromOtherNotary(tx3.Zkpi1s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, true)

		//3添加1,2

		finish, err = l3.AddNewZkpi1FromOtherNotary(tx1.Zkpi1s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, false)

		finish, err = l3.AddNewZkpi1FromOtherNotary(tx2.Zkpi1s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, true)

		tx1, err = l1.StartLockoutStage2()
		assert.EqualValues(t, err, nil)
		tx2, err = l2.StartLockoutStage2()
		assert.EqualValues(t, err, nil)
		tx3, err = l3.StartLockoutStage2()
		assert.EqualValues(t, err, nil)

		//1 添加2,3
		_, finish, err = l1.AddZkpi2(tx2.Zkpi2s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, false)
		ntx1, finish, err := l1.AddZkpi2(tx3.Zkpi2s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, true)

		//2添加1,3
		_, finish, err = l2.AddZkpi2(tx1.Zkpi2s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, false)
		ntx2, finish, err := l2.AddZkpi2(tx3.Zkpi2s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, true)

		//3添加1,2
		_, finish, err = l3.AddZkpi2(tx1.Zkpi2s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, false)
		ntx3, finish, err := l3.AddZkpi2(tx2.Zkpi2s[0])
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, finish, true)

		//1校验签名
		sig1 := l1.CalcSignature(ntx1)
		addr1, err := utils.Ecrecover(utils.Sha3(l1.Message), sig1)
		assert.EqualValues(t, err, nil)
		assert.EqualValues(t, addr1, addr)
		if addr1 != addr {
			cnt++
			continue
		}
		//l1.Message = []byte("bbb")
		//sig1_2 := l1.CalcSignature(nil)
		//addr1_2, err := utils.Ecrecover(utils.Sha3(l1.Message), sig1_2)
		//assert.EqualValues(t, err, nil)
		//assert.EqualValues(t, addr1_2, addr)
		//2校验签名
		//l2.Message = []byte("ccc")
		sig2 := l2.CalcSignature(ntx2)

		addr2, err := utils.Ecrecover(utils.Sha3(l2.Message), sig2)
		assert.EqualValues(t, err, nil)

		//3校验签名
		//l3.Message = []byte("ddd")
		sig3 := l3.CalcSignature(ntx3)

		addr3, err := utils.Ecrecover(utils.Sha3(l3.Message), sig3)
		assert.EqualValues(t, err, nil)
		if addr != addr2 {
			cnt++
			continue
		}
		if addr != addr3 {
			cnt++
			continue
		}
		log.Info(fmt.Sprintf("addr1=%s,addr2=%s,addr3=%s", addr1.String(), addr2.String(), addr3.String()))
	}
	t.Logf("not equal=%d", cnt)
}
