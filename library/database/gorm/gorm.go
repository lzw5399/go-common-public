package fgorm

import (
	"fmt"

	"github.com/lzw5399/go-common-public/library/database/gorm/plugin"
	dbmodel "github.com/lzw5399/go-common-public/library/database/model"
	"github.com/lzw5399/go-common-public/library/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

var DB *gorm.DB

func StartBase(initDbFunc func(config *gorm.Config) (*gorm.DB, error), options ...dbmodel.DBOptionFunc) {
	option := dbmodel.MergeDBOption(options...)

	gormLog := gormLogger.Default.LogMode(log.GormLogModeStrToGormLogLevel(fconfig.DefaultConfig.GormLogMode))
	if fconfig.DefaultConfig.GormLogJson {
		gormLog = log.NewGormJsonLogger(nil)
	}

	// 初始化数据库链接
	var err error
	DB, err = initDbFunc(&gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		panic(fmt.Errorf("start db initDbFunc failed. dbMode(%s) err(%s)", fconfig.DefaultConfig.DBMode, err))
	}

	// 多租户插件
	err = DB.Use(plugin.NewTenantPlugin(option.TenantTableSet))
	if err != nil {
		panic(errors.Wrap(err, "mysql.Start NewTenantPlugin failed"))
	}

	// i18n翻译插件
	err = DB.Use(plugin.NewI18nPlugin(option.I18nTranslators...))
	if err != nil {
		panic(errors.Wrap(err, "mysql.Start NewI18nPlugin failed"))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Errorf("start db get sql db failed. err(%s)", err))
	}
	sqlDB.SetMaxIdleConns(fconfig.DefaultConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(fconfig.DefaultConfig.MaxOpenConns)
}
