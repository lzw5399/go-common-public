package encrypt

import (
	"encoding/hex"

	"github.com/tjfoc/gmsm/sm4"
)

// Sm4Encrypt secret必须是16位, 如果不是16位则需要padding
func Sm4Encrypt(text, secret string) string {
	src := []byte(text)
	key := []byte(secret)

	cipherText, err := sm4.Sm4Ecb(key, src, true)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(cipherText)
}

// Sm4Decrypt secret必须是16位, 如果不是16位则需要padding
func Sm4Decrypt(text, secret string) (string, error) {
	src, err := hex.DecodeString(text)
	if err != nil {
		return "", err
	}
	key := []byte(secret)

	plainText, err := sm4.Sm4Ecb(key, src, false)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
