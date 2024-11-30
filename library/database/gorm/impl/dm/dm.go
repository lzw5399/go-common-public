package dm

import (
	fconfig "github.com/lzw5399/go-common-public/library/config"
	fgorm "github.com/lzw5399/go-common-public/library/database/gorm"
	dbmodel "github.com/lzw5399/go-common-public/library/database/model"
	dm "gitlab.finogeeks.club/finclip-cloud/gorm-driver-dm"
	"gorm.io/gorm"
)

func Start(options ...dbmodel.DBOptionFunc) {
	fgorm.StartBase(func(config *gorm.Config) (*gorm.DB, error) {
		db, err := gorm.Open(dm.Open(fconfig.DefaultConfig.DmURL), config)
		if err != nil {
			return nil, err
		}

		/*err = db.Use(plugin.NewDmPlugin())
		if err != nil {
			panic(errors.Wrap(err, "dm.Start NewDmPlugin failed"))
		}*/

		return db, nil
	}, options...)
}
