package mecdsa

import (
	"os"
	"testing"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, utils.MyStreamHandler(os.Stdout)))
}
func TestKeyGenTNSimple(t *testing.T) {
	_, _, _, _, _, err := KeyGenTN(3, 5)
	if err != nil {
		t.Error(err)
	}

}

func TestKeyGenTN(t *testing.T) {
	_, _, _, _, _, err := KeyGenTN(3, 5)
	if err != nil {
		t.Error(err)
	}
	_, _, _, _, _, err = KeyGenTN(1, 4)
	if err != nil {
		t.Error(err)
	}
	_, _, _, _, _, err = KeyGenTN(3, 5)
	if err != nil {
		t.Error(err)
	}
	_, _, _, _, _, err = KeyGenTN(2, 8)
	if err != nil {
		t.Error(err)
	}
	_, _, _, _, _, err = KeyGenTN(7, 9)
	if err != nil {
		t.Error(err)
	}
	_, _, _, _, _, err = KeyGenTN(1, 4)
	if err != nil {
		t.Error(err)
	}
	_, _, _, _, _, err = KeyGenTN(7, 16)
	if err != nil {
		t.Error(err)
	}
}

func TestSign(t *testing.T) {
	err := Sign(2, 5, 4, []int{4, 1, 2, 3})
	if err != nil {
		t.Error(err)
	}
}
