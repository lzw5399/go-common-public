package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	fconst "github.com/lzw5399/go-common-public/library/context"
	"github.com/lzw5399/go-common-public/library/i18n"
	"github.com/lzw5399/go-common-public/library/util"
)

func Lang(c *gin.Context) {
	lang := c.GetHeader(i18n.HeaderLang)
	if lang == "" { // 为空时，使用默认语言
		lang = fconfig.DefaultConfig.DefaultLang
	}

	// lang必须是配置支持的语言，否则也会降级为默认语言
	matchLang := false
	langs := strings.Split(fconfig.DefaultConfig.SupportLang, ",")
	for _, l := range langs {
		if l == lang {
			matchLang = true
		}
	}

	// 如果入参的lang不在支持的语言列表中，则使用默认语言
	if !matchLang {
		lang = fconfig.DefaultConfig.DefaultLang
	}

	c.Set(i18n.HeaderLang, lang)
	c.Next()
}

func TraceId(c *gin.Context) {
	traceId := c.GetHeader(fconst.HeaderTraceId)
	if traceId == "" {
		traceId = util.NewUUIDString()
	}
	c.Set(fconst.HeaderTraceId, traceId)
	c.Next()
}

func ClientIP(c *gin.Context) {
	ip := util.ClientIP(c)

	c.Set(fconst.HeaderClientIp, ip)
	c.Next()
}

func HttpEndpoint(c *gin.Context) {
	endpoint := fmt.Sprintf("[%s] %s", strings.ToUpper(c.Request.Method), c.Request.URL.Path)

	c.Set(fconst.HeaderHttpEndpoint, endpoint)
	c.Next()
}

func Referer(c *gin.Context) {
	referer := c.GetHeader(fconst.HeaderReferer)

	// 当前的referer去除掉scheme部分，和后面的query参数，只保留host
	refererHost := strings.TrimPrefix(referer, "http://")
	refererHost = strings.TrimPrefix(refererHost, "https://")
	refererHost = strings.Split(refererHost, "/")[0]
	refererHost = strings.Split(refererHost, "?")[0]

	c.Set(fconst.HeaderReferer, refererHost)
	c.Next()
}
