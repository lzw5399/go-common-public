package trace

import (
    "context"
    "fmt"
    "strconv"
    "time"

    "github.com/SkyAPM/go2sky"
    "github.com/SkyAPM/go2sky/propagation"
    v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
    "github.com/gin-gonic/gin"
)

const componentIDGINHttpServer = 5006

type ctxKey struct{}

var ctxKeyInstance = ctxKey{}

// Middleware gin middleware return HandlerFunc  with tracing.
func HttpTraceMiddleware(engine *gin.Engine, tracer *go2sky.Tracer) gin.HandlerFunc {
    if engine == nil || tracer == nil {
        return func(c *gin.Context) {
            c.Next()
        }
    }

    return func(c *gin.Context) {
        span, ctx, err := tracer.CreateEntrySpan(context.Background(), getOperationName(c), func() (string, error) {
            return c.Request.Header.Get(propagation.Header), nil
        })
        if err != nil {
            c.Next()
            return
        }
        span.SetComponent(componentIDGINHttpServer)
        span.Tag(go2sky.TagHTTPMethod, c.Request.Method)
        span.Tag(go2sky.TagURL, c.Request.Host+c.Request.URL.Path)
        span.SetSpanLayer(v3.SpanLayer_Http)

        c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKeyInstance, ctx))

        c.Next()

        if len(c.Errors) > 0 {
            span.Error(time.Now(), c.Errors.String())
        }
        span.Tag(go2sky.TagStatusCode, strconv.Itoa(c.Writer.Status()))
        span.End()
    }
}

func getOperationName(c *gin.Context) string {
    return fmt.Sprintf("/%s%s", c.Request.Method, c.FullPath())
}
