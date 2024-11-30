package trace

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Shopify/sarama"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
)

// noop span, avoid nil
var nSpan = &noopSpan{}

const noTrace = "apm_no_trace"

type Client struct {
	tracer   *go2sky.Tracer
	reporter go2sky.Reporter
	enable   bool

	// http
	httpExtractor func(req *http.Request) propagation.Extractor
	httpInjector  func(req *http.Request) propagation.Injector

	// grpc
	gRpcExtractor func(ctx context.Context) propagation.Extractor
	gRpcInjector  func(ctx *context.Context) propagation.Injector

	//mq:k
	kExtractor func(msg *sarama.ConsumerMessage) propagation.Extractor
	kInjector  func(msg *sarama.ProducerMessage) propagation.Injector
}

func (client *Client) Close() {
	if client.reporter != nil {
		client.reporter.Close()
		client.reporter = nil
	}
	client.tracer = nil
	client.enable = false
}

func (client *Client) NoTraceContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, noTrace, noTrace)
	return ctx
}

func (client *Client) isNoTraceContext(ctx context.Context) bool {
	return ctx == nil || ctx.Value(noTrace) == noTrace
}

func (client *Client) isEnable() bool {
	return client.enable
}

type noopSpan struct {
}

func (b *noopSpan) SetOperationName(string) {
}

func (b *noopSpan) GetOperationName() string {
	return ""
}

func (b *noopSpan) SetPeer(string) {
}

func (b *noopSpan) SetSpanLayer(language_agent.SpanLayer) {
}

func (b *noopSpan) SetComponent(int32) {
}

func (b *noopSpan) Tag(go2sky.Tag, string) {
}

func (b *noopSpan) Log(time.Time, ...string) {
}

func (b *noopSpan) Error(time.Time, ...string) {
}

func (b *noopSpan) End() {
}

func (b *noopSpan) IsEntry() bool {
	return false
}

func (b *noopSpan) IsExit() bool {
	return false
}

func (client *Client) CreateExitSpan(ctx context.Context, operationName string, peer string, injector propagation.Injector) go2sky.Span {
	if !client.enable || client.isNoTraceContext(ctx) {
		return nSpan
	}

	span, err := client.tracer.CreateExitSpan(ctx, operationName, peer, injector)
	if err != nil {
		log.Println(err.Error())
	}

	if span == nil {
		return nSpan
	}

	return span
}

func (client *Client) CreateEntrySpan(ctx context.Context, operationName string, extractor propagation.Extractor) (go2sky.Span, context.Context) {
	if !client.enable || client.isNoTraceContext(ctx) {
		return nSpan, ctx
	}

	span, nCtx, err := client.tracer.CreateEntrySpan(ctx, operationName, extractor)
	if err != nil {
		log.Println(err.Error())
	}

	if span == nil {
		return nSpan, nCtx
	}

	return span, nCtx
}

func (client *Client) CreateLocalSpan(ctx context.Context, opts ...go2sky.SpanOption) (go2sky.Span, context.Context) {
	if !client.enable || client.isNoTraceContext(ctx) {
		return nSpan, ctx
	}

	span, nCtx, err := client.tracer.CreateLocalSpan(ctx, opts...)
	if err != nil {
		log.Println(err.Error())
	}

	if span == nil {
		return nSpan, nCtx
	}

	return span, nCtx
}

func IsNSpan(span go2sky.Span) bool {
	return span == nSpan
}
