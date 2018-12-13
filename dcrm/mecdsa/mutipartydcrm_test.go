package mecdsa

import "testing"

func TestKeyGenTN(t *testing.T) {
	_, _, _, _, err := KeyGenTN(1, 4)
	if err != nil {
		t.Error(err)
	}
}
