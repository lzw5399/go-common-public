package fcontext

import "context"

// ignoreI18nKey 是否在gorm查询的时候，忽略i18n针对某些字段的自动翻译
type ignoreI18nKey struct{}

func IgnoreI18nFromContext(ctx context.Context) bool {
	ignoreI18n, ok := ctx.Value(ignoreI18nKey{}).(bool)
	if !ok {
		return false
	}
	return ignoreI18n
}

func IgnoreI18nWithContext(ctx context.Context, ignoreI18n bool) context.Context {
	return context.WithValue(ctx, ignoreI18nKey{}, ignoreI18n)
}
