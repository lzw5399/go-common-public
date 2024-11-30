package util

import "regexp"

// StrLen 计算字符串长度，中文字符算2个长度，其他字符算1个长度
func StrLen(str string) int {
	pattern := regexp.MustCompile("[\u4e00-\u9fa5]")
	cnt := 0
	for _, char := range str {
		if pattern.MatchString(string(char)) {
			cnt += 2
			continue
		}
		cnt += 1
	}
	return cnt
}
