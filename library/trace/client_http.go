package trace

import (
    "context"
    "errors"
    "net/http"

    "github.com/SkyAPM/go2sky"
    "github.com/SkyAPM/go2sky/propagation"
    "github.com/gin-gonic/gin"
)

func (client *Client) initHttp(config HttpConfig) {
    if config.extractor == nil {
        client.httpExtractor = HttpExtractor
    } else {
        client.httpExtractor = config.extractor
    }
    if config.injector == nil {
        client.httpInjector = HttpInjector
    } else {
        client.httpInjector = config.injector
    }
}

func (client *Client) InjectHttpMiddleware(engine *gin.Engine) {
    if engine == nil {
        panic(errors.New("engine can not be empty"))
    }

    if client.isEnable() {
        engine.Use(HttpTraceMiddleware(engine, client.tracer))
    }
}

func (client *Client) CreateHttpExitSpan(ctx context.Context, req *http.Request, host, path string) go2sky.Span {
    span := client.CreateExitSpan(ctx, path, host, client.httpInjector(req))
    span.Tag(go2sky.TagHTTPMethod, req.Method)
    span.Tag(go2sky.TagURL, host+path)
    return span
}

func (client *Client) CreateHttpExitSpanWithInjector(ctx context.Context, method, host, path string,
    injector propagation.Injector) go2sky.Span {
    span := client.CreateExitSpan(ctx, path, host, injector)
    span.Tag(go2sky.TagHTTPMethod, method)
    span.Tag(go2sky.TagURL, host+path)
    return span
}

func (client *Client) CreateHttpExitSpanWithUrl(ctx context.Context, req *http.Request, url string) go2sky.Span {
    if !client.isEnable() || client.isNoTraceContext(ctx) {
        return nSpan
    }
    _, host, path := ParseURL(url)
    span := client.CreateExitSpan(ctx, path, host, client.httpInjector(req))
    span.Tag(go2sky.TagHTTPMethod, req.Method)
    span.Tag(go2sky.TagURL, host+path)
    return span
}

func (client *Client) CreateHttpExitSpanWithUrlAndInjector(ctx context.Context, method string, url string,
    injector propagation.Injector) go2sky.Span {
    if !client.isEnable() || client.isNoTraceContext(ctx) {
        return nSpan
    }
    _, host, path := ParseURL(url)
    span := client.CreateExitSpan(ctx, path, host, injector)
    span.Tag(go2sky.TagHTTPMethod, method)
    span.Tag(go2sky.TagURL, host+path)
    return span
}

func (client *Client) TraceContextFromGin(c *gin.Context) context.Context {
    if !client.isEnable() || c == nil || c.Request == nil || c.Request.Context() == nil {
        return context.Background()
    }

    res := c.Request.Context().Value(ctxKeyInstance)
    if res == nil {
        return context.Background()
    }

    resCtx, ok := res.(context.Context)
    if !ok {
        return context.Background()
    }
    return resCtx
}

// TraceContextFromGinV2 因为 TraceContextFromGin 中生成的context丢失掉了 gin的Context中的Value。
// 所以添加了V2的方法，保证后续的context能获取到 gin.Context的Value
func (client *Client) TraceContextFromGinV2(c *gin.Context) context.Context {
    ctxNew := context.Background()
    if c == nil || c.Request == nil || c.Request.Context() == nil {
        return ctxNew
    }
    ctxNew = c.Request.Context()

    ctx := &GContext{
        Context: ctxNew,
        gctx:    c,
    }

    if !client.isEnable() {
        return ctx
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
