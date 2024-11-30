package fcontext

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ctxKey struct{}

var ctxKeyInstance = ctxKey{}

// FromGin 因为 保证后续的context能获取到 gin.Context的Value
func FromGin(c *gin.Context) context.Context {
	ctxNew := context.Background()
	if c == nil || c.Request == nil || c.Request.Context() == nil {
		return ctxNew
	}
	ctxNew = c.Request.Context()

	ctx := &GContext{
		Context: ctxNew,
		gctx:    c,
	}

	res := c.Request.Context().Value(ctxKeyInstance)
	if res == nil {
		return ctx
	}

	resCtx, ok := res.(context.Context)
	if !ok {
		return ctx
	}
	return resCtx
}

type GContext struct {
	context.Context
	gctx *gin.Context
}

func (c *GContext) Value(key interface{}) interface{} {
	return c.gctx.Value(key)
}
