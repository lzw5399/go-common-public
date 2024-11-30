package trace

import (
	"context"

	"github.com/SkyAPM/go2sky"
)

func (client *Client) initMongo(config MongoConfig) {

}

func (client *Client) CreateMongoExitSpan(ctx context.Context, action, address, dbName string) go2sky.Span {
    _, host, _ := ParseMongoURLWithCache(address)
    span := client.CreateExitSpan(ctx, action, host, NoopInjector())
    span.Tag(go2sky.TagDBType, DBTypeMongoDB)
    span.Tag(go2sky.TagDBInstance, dbName)
    return span
}
