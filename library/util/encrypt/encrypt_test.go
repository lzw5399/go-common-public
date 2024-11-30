package encrypt

import (
	"fmt"
	"testing"

	"gopkg.in/go-playground/assert.v1"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

func TestCommonEncrypt(t *testing.T) {
	t.Run("sm4", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "sm4"
		fconfig.DefaultConfig.CommonEncryptSm4CryptKey = "finclip9876cloud"
		rawText := "admin"

		// act
		encryptedText := CommonEncrypt(rawText)

		// assert
		assert.Equal(t, encryptedText, "5f6f3f297c60a7eae2a151095d7ba99a")
	})

	t.Run("des", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "des"
		fconfig.DefaultConfig.CommonEncryptDESCryptKey = "w$D5%8x@"
		rawText := "admin"

		// act
		encryptedText := CommonEncrypt(rawText)

		// assert
		assert.Equal(t, encryptedText, "31c6e4a4e60b2e4a")
	})
}

func TestCommonDecrypt(t *testing.T) {
	t.Run("sm4 success", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "sm4"
		fconfig.DefaultConfig.CommonEncryptSm4CryptKey = "finclip9876cloud"
		rawText := "c50cea2d58bef0fad5b8df3c7e5bb66c"

		// act
		decryptedText := CommonDecrypt(rawText)

		// assert
		assert.Equal(t, decryptedText, "admin@open.com")
	})

	t.Run("sm4 failed", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "sm4"
		fconfig.DefaultConfig.CommonEncryptSm4CryptKey = "finclip9876cloud"
		rawText := "947e5836d22131742528c7d1d8f4faf91"

		// act
		decryptedText := CommonDecrypt(rawText)

		// assert
		assert.Equal(t, decryptedText, "")
	})

	t.Run("sm4 empty failed", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "sm4"
		fconfig.DefaultConfig.CommonEncryptSm4CryptKey = "finclip9876cloud"
		rawText := ""

		// act
		decryptedText := CommonDecrypt(rawText)

		// assert
		assert.Equal(t, decryptedText, "")
	})

	t.Run("des success", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "des"
		fconfig.DefaultConfig.CommonEncryptDESCryptKey = "w$D5%8x@"
		rawText := "ff0b1568f335d30b0cb760a867b10d09"

		// act
		decryptedText := CommonDecrypt(rawText)

		// assert
		assert.Equal(t, decryptedText, "13111111111")
	})

	t.Run("des failed", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "des"
		fconfig.DefaultConfig.CommonEncryptDESCryptKey = "w$D5%8x@"
		rawText := "fake_encrypted_str"

		// act
		decryptedText := CommonDecrypt(rawText)

		// assert
		assert.Equal(t, decryptedText, "")
	})

	t.Run("des empty failed", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.CommonEncryptType = "des"
		fconfig.DefaultConfig.CommonEncryptDESCryptKey = "w$D5%8x@"
		rawText := ""

		// act
		decryptedText := CommonDecrypt(rawText)

		// assert
		assert.Equal(t, decryptedText, "")
	})
}

func TestCommonEncryptDecryptByType(t *testing.T) {
	t.Run("encrypt success", func(t *testing.T) {
		// arrange
		rawText := "admin@open.com"

		// act
		encryptedText := CommonEncryptWithTypeAndKey(rawText, "sm4", "finclip9876cloud")

		// assert
		assert.Equal(t, encryptedText, "c50cea2d58bef0fad5b8df3c7e5bb66c")
	})

	t.Run("decrypt success", func(t *testing.T) {
		// arrange
		rawText := "c50cea2d58bef0fad5b8df3c7e5bb66c"

		// act
		decryptedText := CommonDecryptWithTypeAndKey(rawText, "sm4", "finclip9876cloud")

		// assert
		assert.Equal(t, decryptedText, "admin@open.com")
	})
}

func TestGenHex(t *testing.T) {
	decStr := "finclip9876cloud"

	fmt.Println(fmt.Sprintf("%x", decStr))

	// 66696e636c697039383736636c6f7564
}
