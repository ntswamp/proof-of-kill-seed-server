package lib_email

import (
	lib_aes "app/src/lib/aes"
	"crypto/sha512"
	"strings"
)

var cipherKey = []byte("YcabUPAL0=wo$9iE1$%L-,It;wAbr9nR")

func EncryptEmailToDbFormat(mailAddress string) ([]byte, error) {
	addr := strings.ToLower(mailAddress)
	return lib_aes.EncryptWithCFB(cipherKey, addr)
}

func DecryptEmailDbFormat(mailAddress []byte) (string, error) {
	d, err := lib_aes.DecryptWithCFB(cipherKey, mailAddress)
	return string(d), err
}

func HashEmail(mailAddress string) []byte {
	addr := strings.ToLower(mailAddress)
	sum := sha512.Sum512([]byte(addr))
	return sum[:]
}
