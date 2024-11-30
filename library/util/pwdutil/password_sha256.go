package pwdutil

import (
	"crypto/sha256"
	"fmt"

	"github.com/lzw5399/go-common-public/library/util"
)

// genClientPwdFromRawPwdSha256 从 明文密码 生成 客户端传递 的密码
func genClientPwdFromRawPwdSha256(content string) string {
	firstPart := util.Sha256([]byte(util.Sha256([]byte(content))))
	secondPart := util.Sha256([]byte(util.Sha256([]byte(util.Md5([]byte(content))))))
	return firstPart + "_" + secondPart
}

// genDBPwdFromClientPwdSha256 从 客户端传递的密码 生成 数据库存储的密码
func genDBPwdFromClientPwdSha256(clientPwd string, salt string) string {
	if clientPwd == "" {
		return ""
	}
	part1, _ := parseClientPwd(clientPwd)
	return part1
}

// checkClientPwdSha256 对比客户端传递的密码和数据库存储的密码
func checkClientPwdSha256(clientPwd string, dbPwd string, salt string) (isMatch bool, isUpdate bool) {
	part1Pwd, part2Pwd := parseClientPwd(clientPwd)
	if part1Pwd == dbPwd {
		return true, false
	}
	if part2Pwd == sha256Pwd(dbPwd) {
		return true, true
	}
	return false, false
}

func sha256Pwd(pwd string) string {
	firstSha256 := fmt.Sprintf("%x", sha256.Sum256([]byte(pwd)))
	secondSha256 := fmt.Sprintf("%x", sha256.Sum256([]byte(firstSha256)))
	return secondSha256
}
