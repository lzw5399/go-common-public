package transaction

import (
	fgorm "github.com/lzw5399/go-common-public/library/database/gorm"
	"gorm.io/gorm"
)

type GormTransaction struct {
	db *gorm.DB
}

func NewGormTransaction() *GormTransaction {
	return &GormTransaction{
		db: fgorm.DB.Begin(),
	}
}

func (t *GormTransaction) GetTransaction() interface{} {
	return t.db
}

func (t *GormTransaction) Commit() error {
	return t.db.Commit().Error
}

func (t *GormTransaction) Rollback() error {
	return t.db.Rollback().Error
}
