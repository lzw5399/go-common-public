package transaction

import (
	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/log"
)

type ITransaction interface {
	GetTransaction() interface{}
	Commit() error
	Rollback() error
}

func Begin() ITransaction {
	switch fconfig.DefaultConfig.DBMode {
	case fconfig.DB_MODE_MYSQL, fconfig.DB_MODE_DM, fconfig.DB_MODE_GODEN:
		return NewGormTransaction()
	default:
		log.Errorf("transaction.Begin unsupported db mode:(%s), downgrade to dummy transaction", fconfig.DefaultConfig.DBMode)
		return NewDummyTransaction()
	}
}
