package fgrpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	fcontext "github.com/lzw5399/go-common-public/library/context"
	"github.com/lzw5399/go-common-public/library/log"
	fpb "github.com/lzw5399/go-common-public/library/pb"
)

type Ping interface {
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*fpb.PingResp, error)
}

const maxRetry = 60

// EnsureGRPCServiceAlive 因为服务间有依赖，为了确保依赖的服务已经启动，需要在启动时检查依赖的服务是否已经启动
func EnsureGRPCServiceAlive(clients ...Ping) {
	ctx := fcontext.Background()
	for _, cli := range clients {
		for i := 0; i < maxRetry; i++ {
			pingRsp, err := cli.Ping(ctx, &emptypb.Empty{})
			if err == nil {
				log.Warnf("grpc ping (%T) success, response tag: %s", cli, pingRsp.Version)
				break
			}

			if i < maxRetry-1 {
				log.Warnf("grpc ping (%T) failed, retrying (retryCount=%d)", cli, i+1)
				time.Sleep(1 * time.Second)
				continue
			}

			panic(fmt.Errorf("EnsureGRPCServiceAlive grpc ping (%T) failed after max %d retries", cli, maxRetry))
		}
	}
}
