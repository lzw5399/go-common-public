package golden

import (
	fconfig "github.com/lzw5399/go-common-public/library/config"
	fgorm "github.com/lzw5399/go-common-public/library/database/gorm"
	"github.com/lzw5399/go-common-public/library/database/gorm/plugin"
	dbmodel "github.com/lzw5399/go-common-public/library/database/model"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Start(options ...dbmodel.DBOptionFunc) {
	fgorm.StartBase(func(config *gorm.Config) (*gorm.DB, error) {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN: fconfig.DefaultConfig.MysqlURL,
		}), config)
		if err != nil {
			return nil, err
		}
		db.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=1;")

		// golden 字符串默认值处理
		err = db.Use(plugin.NewGoldenDbPlugin())
		if err != nil {
			panic(errors.Wrap(err, "golden.Start NewEmptyStrPlugin failed"))
		}

		return db, nil
	}, options...)
}
