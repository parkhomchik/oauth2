package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/sha3"
)

func EncryptPassword(password string) string {
	h := sha3.New512()
	h.Write([]byte(password))
	b := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(b)
}

func GenerateSecret() string {
	secret := make([]byte, 16)
	rand.Read(secret)
	return fmt.Sprintf("%x", secret)
}
