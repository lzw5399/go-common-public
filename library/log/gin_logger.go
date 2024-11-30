package log

import (
	"github.com/gin-gonic/gin"
)

func GinLogger() gin.HandlerFunc {
	return gin.LoggerWithWriter(defaultLogger.Out)
}
