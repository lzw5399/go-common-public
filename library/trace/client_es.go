package trace

import (
	"context"

	"github.com/SkyAPM/go2sky"
)

func (client *Client) initES(config ESConfig) {

}

func (client *Client) CreateESExitSpan(ctx context.Context, action, address string) go2sky.Span {
    span := client.CreateExitSpan(ctx, action, address, NoopInjector())
    span.Tag(go2sky.TagDBType, DBTypeElasticSearch)
    return span
}
