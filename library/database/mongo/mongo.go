package mongo

import fconfig "github.com/lzw5399/go-common-public/library/config"

// MgoDB 每次使用的时候copy一个，用完之后close
var MgoDB *Session

func Start() {
	if fconfig.DefaultConfig.DBMode != "mongo" {
		return
	}
	url := fconfig.DefaultConfig.MongoURL
	var err error
	MgoDB, err = Dial(url)
	if err != nil {
		panic(err)
	}

	MgoDB.SetMode(Strong, true)
}
