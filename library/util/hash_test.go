package util

import (
	"testing"
)

var str = "hello,world"

func TestMd5(t *testing.T) {
	md5Result := "3cb95cfbe1035bce8c448fcaf80fe7d9"
	if result := Md5([]byte(str)); result != md5Result {
		t.Errorf("md5('hello,world') expected be [%s], but get [%s]", md5Result, result)
	}
}

func TestSha256(t *testing.T) {
	sha256Result := "77df263f49123356d28a4a8715d25bf5b980beeeb503cab46ea61ac9f3320eda"
	if result := Sha256([]byte(str)); result != sha256Result {
		t.Errorf("sha256('hello,world') expected be [%s], but get [%s]", sha256Result, result)
	}
}

func TestSm3(t *testing.T) {
	sm3Result := "72456cdb868a49b85123d6093c15f31c75ac698c466d33d7dc312122f5887d3f"
	if result := Sm3([]byte(str)); result != sm3Result {
		t.Errorf("sm3('hello,world') expected be [%s], but get [%s]", sm3Result, result)
	}
}
