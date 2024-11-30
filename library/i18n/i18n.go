package i18n

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/i18n/plural"
	"gopkg.in/ini.v1"
	"unknwon.dev/i18n"
)

type MessageCode string

var (
	defaultStore      *i18n.Store
	translatorMap     map[Lang]*i18n.Locale
	placeholderRe     = regexp.MustCompile(`\${([a-zA-z]+),\s*(\d+)}`) // e.g. ${file, 1} => ["file", "1"]
	messageReplaceSet = map[string]customDictMapType{}
)

type customDictMapType map[string]map[Lang]string

func init() {
	defaultStore = i18n.NewStore()
	translatorMap = make(map[Lang]*i18n.Locale)
}

func AddLocale(lang Lang, source string) error {
	locale, err := defaultStore.AddLocale(lang.ToLocaleCode(), "", source)
	if err != nil {
		return err
	}
	translatorMap[lang] = locale
	return nil
}

func InitMessageReplaceSet(source string, customDict map[string]map[Lang]string) {
	readI18nKeys(source, customDict)
}

// TDefault 翻译成默认语言
func TDefault(ctx context.Context, code MessageCode, args ...interface{}) string {
	return T(Lang(fconfig.DefaultConfig.DefaultLang), code, args...)
}

// Tc 翻译成context内部传递的语言参数。如果context内没有lang的值，则翻译成默认语言
func Tc(ctx context.Context, code MessageCode, args ...interface{}) string {
	lang := langFromContext(ctx)

	return T(lang, code, args...)
}

// T 翻译成指定语言
func T(lang Lang, code MessageCode, args ...interface{}) string {
	translator, ok := translatorMap[lang]
	if !ok {
		fmt.Printf("locale(%s) not found, downgrade to default lang(%s)\n", lang, fconfig.DefaultConfig.DefaultLang)
		translator = translatorMap[Lang(fconfig.DefaultConfig.DefaultLang)]
	}

	result := translator.Translate(string(code), args...)
	if replaceField, needReplace := messageReplaceSet[string(code)]; needReplace {
		for customDictKey := range replaceField {
			result = strings.ReplaceAll(result, customDictKey, replaceField[customDictKey][lang])
		}
	}
	return result
}

func langFromContext(ctx context.Context) Lang {
	langObj := ctx.Value(HeaderLang)
	if langObj == nil {
		return Lang(fconfig.DefaultConfig.DefaultLang)
	}

	lang, ok := langObj.(string)
	if !ok {
		return Lang(fconfig.DefaultConfig.DefaultLang)
	}

	return Lang(lang)
}

func readI18nKeys(filePath string, customDict customDictMapType) (map[string]customDictMapType, []string, error) {
	file, err := ini.LoadSources(
		ini.LoadOptions{
			IgnoreInlineComment:         true,
			UnescapeValueCommentSymbols: true,
		},
		filePath,
	)
	if err != nil {
		return nil, nil, errors.New("load sources")
	}
	file.BlockMode = false // We only read from the file
	const pluralsSection = "plurals"
	s := file.Section(pluralsSection)
	keys := s.Keys()
	pluralForms := make(map[string]map[plural.Form]string, len(keys))
	for _, k := range s.Keys() {
		fields := strings.SplitN(k.Name(), ".", 2)
		if len(fields) != 2 {
			continue
		}
		noun, form := fields[0], fields[1]
		p, ok := pluralForms[noun]
		if !ok {
			p = make(map[plural.Form]string, 6)
			pluralForms[noun] = p
		}
		switch plural.Form(form) {
		case plural.Zero, plural.One, plural.Two, plural.Few, plural.Many, plural.Other:
			p[plural.Form(form)] = k.String()
		}
	}
	messages := make([]string, 0, len(keys))
	messageReplaceSet = make(map[string]customDictMapType, len(keys))
	for _, s := range file.Sections() {
		if s.Name() == pluralsSection {
			continue
		}
		for _, k := range s.Keys() {
			// NOTE: Majority of messages do not need to deal with plurals, thus it makes
			//  sense to leave them with a nil map to save some memory space.
			var placeholders map[int]*pluralPlaceholder
			format := k.String()
			if strings.Contains(format, "${") {
				matches := placeholderRe.FindAllStringSubmatch(format, -1)
				replaces := make([]string, 0, len(matches)*2)
				for _, submatch := range matches {
					placeholder := submatch[0]
					noun := submatch[1]
					index, _ := strconv.Atoi(submatch[2])
					if index < 1 {
						return nil, nil, errors.New(fmt.Sprintf("the smallest index is 1 but got %d for %q", index, placeholder))
					}
					forms, ok := pluralForms[noun]
					if !ok {
						replaces = append(replaces, placeholder, fmt.Sprintf("<no such plural: %s>", noun))
						continue
					}
					name := fmt.Sprintf("${%d}", index)
					replaces = append(replaces, placeholder, name)
					placeholders[index] = &pluralPlaceholder{
						name:  name,
						forms: forms,
					}
				}
				format = strings.NewReplacer(replaces...).Replace(format)
			}
			key := strings.TrimPrefix(s.Name()+"::"+k.Name(), ini.DefaultSection+"::")
			messages = append(messages, key)

			for customDictKey := range customDict {
				if strings.ContainsAny(k.Value(), customDictKey) {
					messageReplaceSet[key] = customDict
				}
			}
		}
	}
	return messageReplaceSet, messages, nil
}

type pluralPlaceholder struct {
	name  string
	forms map[plural.Form]string
}
