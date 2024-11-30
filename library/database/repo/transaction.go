package repo

import (
	"context"

	fgorm "github.com/lzw5399/go-common-public/library/database/gorm"
	"github.com/lzw5399/go-common-public/library/database/transaction"
	"gorm.io/gorm"
)

func SetTx(tx transaction.ITransaction) RepoOptionFunc {
	return func(option *RepoOption) {
		option.tx = tx
	}
}

func GetGormTx(ctx context.Context, option *RepoOption) *gorm.DB {
	if option == nil {
		return fgorm.DB.WithContext(ctx)
	}

	if option.tx == nil {
		return fgorm.DB.WithContext(ctx)
	}

	tx := option.tx.GetTransaction()
	if tx == nil {
		return fgorm.DB.WithContext(ctx)
	}

	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return fgorm.DB.WithContext(ctx)
	}

	return gormTx.WithContext(ctx)
}
