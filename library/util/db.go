package util

import (
	mgo "github.com/lzw5399/go-common-public/library/database/mongo"
	"gorm.io/gorm"
)

func DbNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound || err == mgo.ErrNotFound
}
