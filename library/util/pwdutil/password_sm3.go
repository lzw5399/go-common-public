package pwdutil

import (
	"fmt"

	"github.com/lzw5399/go-common-public/library/util"
)

// genClientPwdFromRawPwdSm3 从 明文密码 生成 客户端传递 的密码
func genClientPwdFromRawPwdSm3(content string) string {
	firstPart := util.Sm3([]byte(util.Sm3([]byte(content))))
	secondPart := util.Sm3([]byte(util.Sm3([]byte(util.Md5([]byte(content))))))
	return firstPart + "_" + secondPart
}

// genDBPwdFromClientPwdSm3 从 客户端传递的密码 生成 数据库存储的密码
func genDBPwdFromClientPwdSm3(clientPwd string, salt string) string {
	if clientPwd == "" {
		return ""
	}
	part1, _ := parseClientPwd(clientPwd)

	// 数据库存储的要二次加盐哈希
	return util.Sm3WithSalt([]byte(part1), []byte(salt))
}

// checkClientPwdSm3 对比客户端传递的密码和数据库存储的密码
func checkClientPwdSm3(clientPwd string, dbPwd string, salt string) (isMatch bool, isUpdate bool) {
	part1Pwd, _ := parseClientPwd(clientPwd)
	clientDbPwd := util.Sm3WithSalt([]byte(part1Pwd), []byte(salt))
	if clientDbPwd == dbPwd {
		return true, false
	}

	return false, false
}

func sm3Pwd(pwd string) string {
	firstSm3 := fmt.Sprintf("%x", util.Sm3Raw([]byte(pwd)))
	secondSm3 := fmt.Sprintf("%x", util.Sm3Raw([]byte(firstSm3)))
	return secondSm3
}
