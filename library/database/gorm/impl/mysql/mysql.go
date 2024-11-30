package mysql

import (
	fconfig "github.com/lzw5399/go-common-public/library/config"
	fgorm "github.com/lzw5399/go-common-public/library/database/gorm"
	dbmodel "github.com/lzw5399/go-common-public/library/database/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Start(options ...dbmodel.DBOptionFunc) {
	fgorm.StartBase(func(config *gorm.Config) (*gorm.DB, error) {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DriverName: string(fconfig.DefaultConfig.DBMode),
			DSN:        fconfig.DefaultConfig.MysqlURL, // data source name, refer https://github.com/go-sql-driver/mysql#dsn-data-source-name
		}), config)
		if err != nil {
			return nil, err
		}
		db.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=1;")

		return db, nil
	}, options...)
}
