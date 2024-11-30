package pwdutil

import (
	"strings"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

// GenClientPwdFromRawPwd 从 明文密码 生成 客户端传递 的密码
func GenClientPwdFromRawPwd(content string) string {
	cfg := fconfig.DefaultConfig
	switch cfg.PwdHashType {
	case "sm3":
		return genClientPwdFromRawPwdSm3(content)
	default:
		return genClientPwdFromRawPwdSha256(content)
	}
}

// GenDBPwdFromClientPwd 从 客户端传递的密码 生成 数据库存储的密码
func GenDBPwdFromClientPwd(clientPwd string, salt string) string {
	cfg := fconfig.DefaultConfig
	switch cfg.PwdHashType {
	case "sm3":
		return genDBPwdFromClientPwdSm3(clientPwd, salt)
	default:
		return genDBPwdFromClientPwdSha256(clientPwd, salt)
	}
}

// GenDBPwdFromRawPwd 从 明文密码 生成 数据库存储的密码
func GenDBPwdFromRawPwd(rawPwd string, salt string) string {
	cfg := fconfig.DefaultConfig
	switch cfg.PwdHashType {
	case "sm3":
		return genDBPwdFromClientPwdSm3(genClientPwdFromRawPwdSm3(rawPwd), salt)
	default:
		return genDBPwdFromClientPwdSha256(genClientPwdFromRawPwdSha256(rawPwd), salt)
	}
}

// CheckClientPwd 对比客户端传递的密码和数据库存储的密码
func CheckClientPwd(clientPwd string, dbPwd string, salt string) (isMatch bool, isUpdate bool) {
	cfg := fconfig.DefaultConfig
	switch cfg.PwdHashType {
	case "sm3":
		return checkClientPwdSm3(clientPwd, dbPwd, salt)
	default:
		return checkClientPwdSha256(clientPwd, dbPwd, salt)
	}
}

func parseClientPwd(clientPwd string) (string, string) {
	ss := strings.Split(clientPwd, "_")
	if len(ss) == 2 {
		return ss[0], ss[1]
	} else {
		return ss[0], ""
	}
}
