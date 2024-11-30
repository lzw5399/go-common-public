package util

import (
	"io/ioutil"
	"os"

	"github.com/google/uuid"
)

func CryptoFile(fileBytes []byte, appId string) (outBytes []byte, err error) {
	hexStr := Md5([]byte("miniprogram" + appId))
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	uuidStr := uuid.New().String()
	path := dir + "/" + uuidStr + "/"
	zipFilePath := dir + "/" + uuidStr + ".zip"
	err = UnZip(fileBytes, path)
	if err != nil {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	err = Compress([]*os.File{f}, hexStr, zipFilePath, uuidStr)
	defer func() {
		_ = os.RemoveAll(path)
		_ = os.Remove(zipFilePath)
	}()
	if err != nil {
		return
	}

	outBytes, err = ioutil.ReadFile(zipFilePath)
	return
}
