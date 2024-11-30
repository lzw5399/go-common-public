package fmq

import (
	"context"
	"sync"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/mq/k"
	"github.com/lzw5399/go-common-public/library/mq/n"
)

var (
	nOnce sync.Once
)

// ProduceMessage 生产消息
func ProduceMessage(ctx context.Context, topic string, msg []byte) error {
	cfg := fconfig.DefaultConfig
	switch cfg.MQMode {
	case "n":
		nOnce.Do(n.Init)
		return n.SendNMsg(ctx, topic, msg)
	case "k":
		return k.SendKMsg(ctx, topic, string(msg))
	}

	return nil
}

// RegisterConsumerCallback 注册消费者回调
func RegisterConsumerCallback(ctx context.Context, topic string, group string, handler func(ctx context.Context, msg []byte) error) {
	cfg := fconfig.DefaultConfig
	switch cfg.MQMode {
	case "n":
		nOnce.Do(n.Init)
		n.QueueSub(ctx, topic, group, handler)
	case "k":
		go k.StartKClusterClient(cfg.KAddr, topic, group, handler)
	}
}
