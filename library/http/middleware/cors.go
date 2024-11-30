package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	fconfig "github.com/lzw5399/go-common-public/library/config"
)

func CORS(c *gin.Context) {
	allowOrigins := fconfig.DefaultConfig.CORSAllowOrigins

	c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigins)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PUT, DELETE, UPDATE")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
