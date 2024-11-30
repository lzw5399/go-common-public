package database

import (
	dbmodel "github.com/lzw5399/go-common-public/library/database/model"
	"github.com/lzw5399/go-common-public/library/database/mongo"
	"github.com/lzw5399/go-common-public/library/log"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/database/gorm/impl/dm"
	"github.com/lzw5399/go-common-public/library/database/gorm/impl/golden"
	"github.com/lzw5399/go-common-public/library/database/gorm/impl/mysql"
)

func Start(options ...dbmodel.DBOptionFunc) {
	switch fconfig.DefaultConfig.DBMode {
	case fconfig.DB_MODE_MYSQL:
		mysql.Start(options...)
	case fconfig.DB_MODE_DM:
		dm.Start(options...)
	case fconfig.DB_MODE_GODEN:
		golden.Start(options...)
	case fconfig.DB_MODE_MONGO:
		mongo.Start()
	default:
		log.Infof("database Start current db_mode(%s) fallback to default mode: mysql", fconfig.DefaultConfig.DBMode)
		mysql.Start(options...)
	}
}
