package trace

import (
	"context"

	"github.com/SkyAPM/go2sky/propagation"
	"google.golang.org/grpc/metadata"
)

func GRpcExtractor(ctx context.Context) propagation.Extractor {
    return func() (s string, e error) {
        // https://github.com/googleapis/google-cloud-go/issues/624
        // After some recent changes in grpc-go,
        // metadata.NewContext now assumes you are accessing the outgoing metadata,
        // and metadata.FromContext assumes you are retrieving the incoming credentials.
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return "", nil
        }
        header, ok := md[propagation.Header]
        if !ok {
            return "", nil
        }
        if len(header) <= 0 {
            return "", nil
        }
        return header[0], nil
    }
}

func GRpcInjector(ctx *context.Context) propagation.Injector {
    return func(header string) error {
        md := metadata.New(map[string]string{propagation.Header: header})
        newCtx := metadata.NewOutgoingContext(*ctx, md)
        *ctx = newCtx
        return nil
    }
}
