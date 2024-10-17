package helper

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func RandomString(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func ExtractExtension(filename string) string {
	parts := strings.Split(filename, ".")
	return parts[len(parts)-1]
}
