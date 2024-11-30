package middleware

import "github.com/gin-gonic/gin"

func Transfer(c *gin.Context) {
    userId := c.Request.Header.Get("x-mop-fcid")
    if userId != "" {
        c.Request.Header.Set("X-Consumer-Custom-ID", userId)
    }
    c.Next()
}
