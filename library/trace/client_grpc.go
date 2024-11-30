package trace

import (
	"context"

	"github.com/SkyAPM/go2sky"
)

func (client *Client) initGRpc(config GRpcConfig) {
    if config.extractor == nil {
        client.gRpcExtractor = GRpcExtractor
    } else {
        client.gRpcExtractor = config.extractor
    }

    if config.injector == nil {
        client.gRpcInjector = GRpcInjector
    } else {
        client.gRpcInjector = config.injector
    }
}

func (client *Client) CreateGRpcEntrySpan(ctx context.Context, method string) (go2sky.Span, context.Context) {
    span, nCtx := client.CreateEntrySpan(context.Background(), method, client.gRpcExtractor(ctx))
    span.Tag(TagGRpcMethod, method)
    return span, nCtx
}

func (client *Client) CreateGRpcExitSpan(ctx context.Context, operationName, method, thirdService string) (go2sky.Span, context.Context) {
    nCtx := ctx
    span := client.CreateExitSpan(ctx, operationName, thirdService, client.gRpcInjector(&nCtx))
    span.Tag(TagGRpcMethod, method)
    return span, nCtx
}
