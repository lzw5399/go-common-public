package util

import (
	"github.com/gin-gonic/gin"

	"github.com/lzw5399/go-common-public/library/log"
)

func ClientIP(c *gin.Context) string {
	//xRealIp := strings.TrimSpace(c.Request.Header.Get("X-Real-Ip"))
	//if xRealIp != "" {
	//	log.Debugf("[ClientIP] X-Real-Ip: %s", xRealIp)
	//	return xRealIp
	//}

	//clientIPStr := c.Request.Header.Get("X-Forwarded-For")
	//clientIP := strings.TrimSpace(strings.Split(clientIPStr, ",")[0])
	//if clientIP != "" {
	//	log.Debugf("[ClientIP] X-Forwarded-For: %s", clientIP)
	//	return clientIP
	//}

	ip := c.ClientIP()
	log.Debugf("[ClientIP] ClientIP: %s", ip)
	return ip
}
