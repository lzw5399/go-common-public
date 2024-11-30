package util

import (
	"crypto/md5"
	"encoding/hex"
)

func GenMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Get16MD5Encode(data string) string {
	return GenMd5(data)[8:24]
}
