package lib_email

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func GenerateAuthCode() (string, error) {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
