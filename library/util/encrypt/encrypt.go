package encrypt

import fconfig "github.com/lzw5399/go-common-public/library/config"

func CommonEncrypt(text string) string {
	cfg := fconfig.DefaultConfig
	switch cfg.CommonEncryptType {
	case "sm4":
		return Sm4Encrypt(text, cfg.CommonEncryptSm4CryptKey)
	default:
		return EncryptDES_ECB(text, cfg.CommonEncryptDESCryptKey)
	}
}

func CommonDecrypt(text string) string {
	if text == "" {
		return ""
	}

	cfg := fconfig.DefaultConfig
	var (
		encryptedText string
		err           error
	)

	switch cfg.CommonEncryptType {
	case "sm4":
		encryptedText, err = Sm4Decrypt(text, cfg.CommonEncryptSm4CryptKey)
	default:
		encryptedText, err = DecryptDES_ECB_V2(text, cfg.CommonEncryptDESCryptKey)
	}

	if err != nil {
		return ""
	}
	return encryptedText
}

func CommonEncryptWithTypeAndKey(text, encryptType, key string) string {
	switch encryptType {
	case "sm4":
		return Sm4Encrypt(text, key)
	default:
		return EncryptDES_ECB(text, key)
	}
}

func CommonDecryptWithTypeAndKey(text, encryptType, key string) string {
	if text == "" {
		return ""
	}

	var (
		encryptedText string
		err           error
	)
	switch encryptType {
	case "sm4":
		encryptedText, err = Sm4Decrypt(text, key)
	default:
		encryptedText, err = DecryptDES_ECB_V2(text, key)
	}
	if err != nil {
		return ""
	}

	return encryptedText
}
