package lib_aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
)

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

func Encrypt(plaintext string) (string, []byte, error) {
	var key string
	var ciphertext []byte
	// Generate a random set of 32 bytes for AES-256
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	if err != nil {
		return key, ciphertext, err
	}

	// Encode key to string
	key = hex.EncodeToString(keyBytes)

	// Create new cipher block using the 32byte key
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return key, ciphertext, err
	}

	// Wrap block in GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return key, ciphertext, err
	}

	// create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return key, ciphertext, err
	}

	// Seal the address
	ciphertext = aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	return key, ciphertext, nil
}

func EncryptGCMUsingKey(plaintext, key string) ([]byte, error) {
	var ciphertext []byte
	// Generate a random set of 32 bytes for AES-256
	keyBytes := []byte(key)

	// Create new cipher block using the 32byte key
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return ciphertext, err
	}

	// Wrap block in GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertext, err
	}

	// create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return ciphertext, err
	}

	// Seal the address
	ciphertext = aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

func Decrypt(keyString string, data []byte) (string, error) {
	var plaintext string
	// turn keyString back into byte array
	key, _ := hex.DecodeString(keyString)

	// get cipher block for this key
	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext, err
	}

	// wrap in gcm
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext, err
	}

	// get noncesize and split the data into nonce and ciphertext parts
	nonceSize := aesGCM.NonceSize()
	nonce, cipherText := data[:nonceSize], data[nonceSize:]

	// open the seal
	decryptedText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return plaintext, err
	}
	plaintext = string(decryptedText)
	return plaintext, nil
}

func EncryptWithCFB(key []byte, plaintext string) ([]byte, error) {
	var cipherText []byte
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return cipherText, err
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return cipherText, err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cipherText = make([]byte, len(plaintext))
	cfb.XORKeyStream(cipherText, []byte(plaintext))
	data := make([]byte, len(cipherText)+16)
	copy(data, iv)
	copy(data[16:], cipherText)
	return data, nil
}

func DecryptWithCFB(key []byte, data []byte) ([]byte, error) {
	iv, encryptedText := data[:16], data[16:]
	var plaintext = make([]byte, len(encryptedText))
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return plaintext, err
	}

	cfbdec := cipher.NewCFBDecrypter(block, iv)
	cfbdec.XORKeyStream(plaintext, encryptedText)
	return plaintext, nil
}
