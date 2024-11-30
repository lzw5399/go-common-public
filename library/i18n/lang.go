package i18n

const (
	HeaderLang = "lang"
)

type Lang string

const (
	LangZh   Lang = "zh"
	LangEn   Lang = "en"
	LangZhHk Lang = "zh-HK"
)

func (l Lang) String() string {
	return string(l)
}

func (l Lang) ToLocaleCode() string {
	code, ok := langToLocaleCode[l]
	if !ok {
		return ""
	}
	return code
}

var (
	langToLocaleCode = map[Lang]string{
		LangZh:   "zh-CN",
		LangEn:   "en-US",
		LangZhHk: "zh-HK",
	}
)
