package utils

import (
	"encoding/base64"

	"golang.org/x/crypto/sha3"
)

func EncryptPassword(password string) string {
	h := sha3.New512()
	h.Write([]byte(password))
	b := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(b)
}
