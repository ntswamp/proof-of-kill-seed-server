package lib_aes

import (
	"testing"
)

func TestCFB(t *testing.T) {
	cipherKey := []byte("Xkfwkw9ArDJy9WNq8hEuLbjZb5GFctuX")
	testText := "randomtesttext"

	cipherText1, err := EncryptWithCFB(cipherKey, testText)
	if err != nil {
		t.Error(err)
	}
	cipherText2, err := EncryptWithCFB(cipherKey, testText)
	if err != nil {
		t.Error(err)
	}
	if string(cipherText1) == string(cipherText2) {
		t.Errorf("Encrypt function returned same data twice\n")
	}
	decData1, err := DecryptWithCFB(cipherKey, cipherText1)
	if err != nil {
		t.Error(err)
	}
	decData2, err := DecryptWithCFB(cipherKey, cipherText2)
	if err != nil {
		t.Error(err)
	}
	if string(decData1) != string(decData2) {
		t.Errorf("Decrypt function did not return same decrypted data\n")
	}
}
