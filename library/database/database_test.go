package database

import (
	"context"
	"testing"

	"github.com/lzw5399/go-common-public/library/database/repo"
	"github.com/lzw5399/go-common-public/library/database/transaction"

	"github.com/lzw5399/go-common-public/library/database/gorm/impl/mysql"
)

func TestTransaction(t *testing.T) {
	transaction.Begin()
}

func TestMysql(t *testing.T) {
	mysql.Start()
}

func TestRepo(t *testing.T) {
	repo.GetGormTx(context.Background(), nil)
}
