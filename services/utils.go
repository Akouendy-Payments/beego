package services

import (
	"crypto/sha512"
	"encoding/hex"
)

func Hash512(text string) string {
	hasher := sha512.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
