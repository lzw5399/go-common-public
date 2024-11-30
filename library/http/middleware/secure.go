package middleware

import (
	"net"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	fconfig "github.com/lzw5399/go-common-public/library/config"
	ferrors "github.com/lzw5399/go-common-public/library/errors"
	"github.com/lzw5399/go-common-public/library/http/httputil"
)

// Secure is a middleware function that appends security
// and resource access headers.
func Secure(c *gin.Context) {
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")
	if c.Request.TLS != nil {
		c.Header("Strict-Transport-Security", "max-age=31536000")
	}
}

func RefererWhitelistCheck(c *gin.Context) {
	cfg := fconfig.DefaultConfig

	// 如果referer允许所有域名，则直接跳过
	if cfg.RefererAllowDomains == "*" {
		c.Next()
		return
	}

	// 获取referer
	referer := c.GetHeader("Referer")
	if referer == "" {
		c.Next()
		return
	}

	parsedURL, err := url.Parse(referer)
	if err != nil {
		httputil.MakeRspWithRspInfo(c, ferrors.Forbidden(), gin.H{})
		return
	}

	// 判断referer是否在允许的域名列表中
	parsedReferer := strings.ToLower(parsedURL.Scheme) + "://" + strings.ToLower(parsedURL.Host)
	_, ok := cfg.RefererAllowDomainSet[parsedReferer]
	if !ok {
		httputil.MakeRspWithRspInfo(c, ferrors.Forbidden(), gin.H{})
		return
	}

	c.Next()
	return
}

func XForwardedForCheck(c *gin.Context) {
	xff := c.Request.Header.Get("X-Forwarded-For")
	if xff == "" || fconfig.DefaultConfig.XForwardedForAllowNetCIDR == "*" {
		c.Next()
		return
	}

	allowCIDRArr := fconfig.DefaultConfig.XForwardedForAllowNetCIDRArr
	if len(allowCIDRArr) == 0 {
		c.Next()
		return
	}

	isAllowed := xForwardedForCheck(xff, allowCIDRArr)
	if !isAllowed {
		httputil.MakeRspWithRspInfo(c, ferrors.Forbidden(), gin.H{})
		return
	}

	c.Next()
	return
}

func xForwardedForCheck(xff string, allowCIDRArr []*net.IPNet) bool {
	if xff == "" {
		return true
	}

	arr := strings.Split(xff, ",")
	for _, ipStr := range arr {
		// 解析ip
		ip := net.ParseIP(strings.TrimSpace(ipStr))
		if ip == nil {
			return false
		}

		// 检查是否在允许的网段内
		isAllowed := false
		for _, cidr := range allowCIDRArr {
			if cidr.Contains(ip) {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return false
		}
	}

	return true
}
