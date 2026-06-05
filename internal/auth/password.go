package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashPassword(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := sha256.Sum256(append(salt, []byte(password)...))
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash[:])
}

func VerifyPassword(password, hashed string) bool {
	parts := strings.SplitN(hashed, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	hash := sha256.Sum256(append(salt, []byte(password)...))
	return hex.EncodeToString(hash[:]) == parts[1]
}
