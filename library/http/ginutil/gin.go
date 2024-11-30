package ginutil

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/http/middleware"
	"github.com/lzw5399/go-common-public/library/log"
)

func NewGinDefault() *gin.Engine {
	cfg := fconfig.DefaultConfig

	g := gin.New()
	g.Use(log.GinLogger())
	g.Use(middleware.Recovery())
	if len(cfg.TrustedProxiesCIDRArr) > 0 {
		err := g.SetTrustedProxies(cfg.TrustedProxiesCIDRArr)
		if err != nil {
			panic(errors.Wrap(err, "NewGinDefault SetTrustedProxies failed"))
		}
	}

	return g
}
