package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"

	"github.com/tjfoc/gmsm/sm3"
)

// Md5 hash算法
func Md5(data []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Sha256 hash算法
func Sha256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Sm3 国密hash算法
func Sm3(data []byte) string {
	h := sm3.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Sm3Raw 国密hash算法
func Sm3Raw(data []byte) []byte {
	h := sm3.New()
	h.Write(data)
	return h.Sum(nil)
}

func Sm3WithSalt(data []byte, salt []byte) string {
	h := sm3.New()
	h.Write(data)
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}

func Sm3RawWithSalt(data []byte, salt []byte) []byte {
	h := sm3.New()
	h.Write(data)
	h.Write(salt)
	return h.Sum(nil)
}

func FHash(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func FHashRaw(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

func FHashWithSalt(data []byte, salt []byte) string {
	h := md5.New()
	h.Write(data)
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}

func FHashRawWithSalt(data []byte, salt []byte) []byte {
	h := md5.New()
	h.Write(data)
	h.Write(salt)
	return h.Sum(nil)
}
