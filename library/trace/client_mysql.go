package trace

import (
	"context"

	"github.com/SkyAPM/go2sky"
)

func (client *Client) initMySql(config MySqlConfig) {

}

func (client *Client) CreateMySqlExitSpan(ctx context.Context, action, address, dbName string) go2sky.Span {
    _, host, _ := ParseURLWithCache(address)
    span := client.CreateExitSpan(ctx, action, host, NoopInjector())
    span.Tag(go2sky.TagDBType, DBTypeMySQL)
    span.Tag(go2sky.TagDBInstance, dbName)
    return span
}
