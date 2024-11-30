package util

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	DES_CRYPT_KEY = "w$D5%8x@"
	DES_CRYPT_IV  = "34RHn876"
)

type DES_CBC struct {
}

// 分组填充--填充对应长度的相同数字（1， 22， 333....）
//data: 待分组数据, blockSize: 每组的长度
func (d DES_CBC) PKCS7Padding(date []byte, blockSize int) (PaddingResult []byte) {
	// 获取数据长度
	length := len(date)
	// 获取待填充数据长度
	count := length % blockSize
	PaddingCount := blockSize - count
	// 在数据后填充数据
	PaddingDate := bytes.Repeat([]byte{byte(PaddingCount)}, PaddingCount)
	PaddingResult = append(date, PaddingDate...)
	return
}

//  分组移除
func (d DES_CBC) PKCS7Unpadding(date []byte, blockSize int) (UnpaddingResult []byte) {
	length := len(date)
	temp := int(date[length-1])
	UnpaddingResult = date[:length-temp]
	return
}

//src:           --明文/密文，需要分组填充，每组8byte
//key:           --秘钥   8byte
//iv:            --初始化向量  8byte  长度必须与key相同
//加密
func (d DES_CBC) DesCBCEncrypt(src []byte, key []byte, iv []byte) []byte {
	// 创建并返回一个使用DES算法的cipher.Block接口
	block, err := des.NewCipher(key)
	// 判断是否创建成功
	if err != nil {
		panic(err)
	}
	// 明文组数据填充
	paddingText := d.PKCS7Padding(src, block.BlockSize())
	// 创建一个密码分组为链接模式的, 底层使用DES加密的BlockMode接口
	blockMode := cipher.NewCBCEncrypter(block, iv)
	// 加密
	dst := make([]byte, len(paddingText))
	blockMode.CryptBlocks(dst, paddingText)
	return dst
}

// 解密：
func (d DES_CBC) DesCBCDecrypt(src []byte, key []byte, iv []byte) []byte {
	// 创建并返回一个使用DES算法的cipher.Block接口
	block, err := des.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// 创建一个密码分组为链接模式的, 底层使用DES解密的BlockMode接口
	blockMode := cipher.NewCBCDecrypter(block, iv)
	// 解密
	dst := make([]byte, len(src))
	blockMode.CryptBlocks(dst, src)
	// 分组移除
	dst = d.PKCS7Unpadding(dst, block.BlockSize())
	return dst
}

type DES_ECB struct {
}

//ECB加密
func (d DES_ECB) EncryptDES_ECB(src, key string) string {
	data := []byte(src)
	keyByte := []byte(key)
	block, err := des.NewCipher(keyByte)
	if err != nil {
		panic(err)
	}
	bs := block.BlockSize()
	//对明文数据进行补码
	data = d.PKCS5Padding(data, bs)
	if len(data)%bs != 0 {
		panic("Need a multiple of the blocksize")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		//对明文按照blocksize进行分块加密
		//必要时可以使用go关键字进行并行加密
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	return fmt.Sprintf("%x", out)
}

func (d DES_ECB) DecryptDES_ECB_SAFE(src, key string) (string, error) {
	data, err := hex.DecodeString(src)
	if err != nil {
		//fmt.Println(err)
		// panic(err)
		return "", err
	}
	keyByte := []byte(key)
	block, err := des.NewCipher(keyByte)
	if err != nil {
		// panic(err)
		return "", err
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		// panic("crypto/cipher: input not full blocks")
		return "", err
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = d.PKCS5UnPadding(out)
	return string(out), nil
}

//ECB解密
func (d DES_ECB) DecryptDES_ECB(src, key string) string {
	data, err := hex.DecodeString(src)
	if err != nil {
		//fmt.Println(err)
		//panic(err)
		return src
	}
	keyByte := []byte(key)
	block, err := des.NewCipher(keyByte)
	if err != nil {
		//panic(err)
		return src
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		//panic("crypto/cipher: input not full blocks")
		return src
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = d.PKCS5UnPadding(out)
	return string(out)
}

//ECB解密
func (d DES_ECB) DecryptDES_ECB_V2(src, key string) (result string, err error) {
	defer func() {
		if defErr := recover(); defErr != nil {
			fmt.Printf("DecryptDES_ECB_V2 panic info:%v\n", defErr)
			result = ""
			err = errors.New("DecryptDES_ECB_V2 err")
		}
	}()
	data, err := hex.DecodeString(src)
	if err != nil {
		return "", err
	}
	keyByte := []byte(key)
	block, err := des.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return "", errors.New("crypto/cipher: input not full blocks")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = d.PKCS5UnPadding(out)
	return string(out), nil
}

//明文补码算法
func (d DES_ECB) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//明文减码算法
func (d DES_ECB) PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		tmpStr := ""
		return []byte(tmpStr)
	}
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
