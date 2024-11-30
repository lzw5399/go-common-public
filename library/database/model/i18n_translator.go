package dbmodel

import (
	"context"

	"github.com/lzw5399/go-common-public/library/i18n"
)

type FieldI18nTranslator interface {
	TableName() string
	Translate(ctx context.Context, lang i18n.Lang, originDest interface{}) (dest interface{})
}
