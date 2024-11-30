package i18n

import (
	"fmt"
	"testing"
)

func TestI18n(t *testing.T) {
	fmt.Println(T(LangZh, "role.admin"))
}
