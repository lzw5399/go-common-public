package plugin

import (
	fconfig "github.com/lzw5399/go-common-public/library/config"
	fcontext "github.com/lzw5399/go-common-public/library/context"
	dbmodel "github.com/lzw5399/go-common-public/library/database/model"
	"github.com/lzw5399/go-common-public/library/i18n"
	"gorm.io/gorm"
)

const (
	_eventAfterQuery = "i18n:after_query"
)

type I18nPlugin struct {
	tableNameToTranslator map[string]dbmodel.FieldI18nTranslator
}

func NewI18nPlugin(
	translators ...dbmodel.FieldI18nTranslator,
) gorm.Plugin {
	tableNameToTranslator := make(map[string]dbmodel.FieldI18nTranslator)
	for _, translator := range translators {
		_, exist := tableNameToTranslator[translator.TableName()]
		if exist {
			panic("NewI18nPlugin duplicate translator for table: " + translator.TableName())
		}

		tableNameToTranslator[translator.TableName()] = translator
	}

	return &I18nPlugin{
		tableNameToTranslator: tableNameToTranslator,
	}
}

func (i *I18nPlugin) Name() string {
	return "I18nPlugin"
}

func (i *I18nPlugin) Initialize(db *gorm.DB) (err error) {
	// Register various callback events in GORM
	for _, e := range []error{
		db.Callback().Query().After("gorm:query").Register(_eventAfterQuery, i.afterQuery),
	} {
		if e != nil {
			return e
		}
	}
	return
}

func (i *I18nPlugin) afterQuery(db *gorm.DB) {
	if db == nil || db.Statement == nil || db.Statement.Context == nil || db.Statement.Table == "" {
		return
	}

	ctx := db.Statement.Context

	// 如果不是指定表的话，直接跳过，不进行i18n翻译
	translator, ok := i.tableNameToTranslator[db.Statement.Table]
	if !ok {
		return
	}

	// 判断是否手动跳过i18n翻译条件
	ignoreI18n := fcontext.IgnoreI18nFromContext(ctx)
	if ignoreI18n {
		return
	}

	// 如果当前上下文是通过默认语言查询的话，就不用进行翻译
	currentLang := fcontext.LangFromContext(ctx)
	defaultLang := i18n.Lang(fconfig.DefaultConfig.DefaultLang)
	isQueryByDefaultLang := currentLang == defaultLang
	if isQueryByDefaultLang {
		return
	}

	// 翻译并替换原始返回
	db.Statement.Dest = translator.Translate(ctx, currentLang, db.Statement.Dest)
	db.Statement.Model = db.Statement.Dest
}
