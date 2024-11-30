package n

import (
	"context"
	"sync"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/pkg/errors"
)

var (
	nClient *NClient
	nOnce   sync.Once
)

func Init() {
	nOnce.Do(func() {
		cfg := fconfig.DefaultConfig
		if cfg.NUrl == "" || cfg.MQMode != "n" {
			return
		}
		nClient = NewNClient(cfg.NUrl)
		nClient.Start()
	})
}

func QueueSub(ctx context.Context, topic, queue string, cb ConsumerCb) {
	nClient.QueueSub(ctx, topic, queue, cb)
}

func SendNMsg(ctx context.Context, topic string, msg []byte) error {
	if nClient == nil {
		return errors.New("MQ:n not Init")
	}
	return nClient.Pub(ctx, topic, msg)
}
