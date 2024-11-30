package trace

import (
	"context"

	"github.com/SkyAPM/go2sky"
)

func (client *Client) initRedis(config RedisConfig) {

}

func (client *Client) CreateRedisExitSpan(ctx context.Context, action, address, dbName string) go2sky.Span {
    span := client.CreateExitSpan(ctx, action, address, NoopInjector())
    span.Tag(go2sky.TagDBType, DBTypeRedis)
    span.Tag(go2sky.TagDBInstance, dbName)
    return span
}
