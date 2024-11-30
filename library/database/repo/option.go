package repo

import (
	"context"

	fcontext "github.com/lzw5399/go-common-public/library/context"
	"github.com/lzw5399/go-common-public/library/database/transaction"
)

type RepoOptionFunc func(*RepoOption)

type RepoOption struct {
	tx           transaction.ITransaction
	ignoreTenant bool // 是否忽略租户限制
	ignoreI18n   bool // 是否在gorm查询的时候，跳过i18n针对某些字段的自动翻译
}

func MergeRepoOption(ctx context.Context, opts ...RepoOptionFunc) (context.Context, *RepoOption) {
	option := &RepoOption{}
	for _, opt := range opts {
		opt(option)
	}

	ctx = fcontext.IgnoreTenantWithContext(ctx, option.ignoreTenant)
	ctx = fcontext.IgnoreI18nWithContext(ctx, option.ignoreI18n)

	return ctx, option
}
