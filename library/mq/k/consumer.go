package k

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/log"
	"github.com/lzw5399/go-common-public/library/trace"
)

type ConsumerCallbackHandler func(ctx context.Context, input []byte) error // 使用者需要自定义该回调函数实现体

func StartKClusterClient(kAddr, topic, group string, handler ConsumerCallbackHandler) {
	version, err := sarama.ParseKafkaVersion(fconfig.DefaultConfig.KVersion)
	if err != nil {
		log.Errorf("mq:k err: %s", err)
		return
	}
	cfg := cluster.NewConfig()
	cfg.Consumer.Return.Errors = true
	cfg.Group.Return.Notifications = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest // OffsetOldest OffsetNewest
	cfg.Consumer.Offsets.CommitInterval = 1 * time.Second
	cfg.Version = version

	if fconfig.DefaultConfig.KUser != "" && fconfig.DefaultConfig.KPwd != "" {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = fconfig.DefaultConfig.KUser
		cfg.Net.SASL.Password = fconfig.DefaultConfig.KPwd
	}

	var consumer *cluster.Consumer
	for {
		var err error
		consumer, err = cluster.NewConsumer(strings.Split(kAddr, ","), group, strings.Split(topic, ","), cfg)
		if err != nil {
			log.Errorf("Failed to start k consumer,err: %s", err)
			time.Sleep(time.Millisecond * 5000)
			continue
		} else {
			break
		}
	}

	defer consumer.Close()

	// Create signal channel
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	log.Debugf("Sarama consumer up and running!...")

	// Consume all channels, wait for signal to exit
	for {
		select {
		case msg, more := <-consumer.Messages():
			if more {
				func() {
					span, ctx := trace.ApmClient().CreateKEntrySpan(context.Background(), msg.Topic, "handler", msg)
					defer span.End()
					consumer.MarkOffset(msg, "")
					log.Debugf("Message claimed: value = %s, timestamp = %v, topic = %s", string(msg.Value), msg.Timestamp, msg.Topic)
					if err := handler(ctx, msg.Value); err != nil {
						log.Errorf("Callback handler err: %s", err)
					}
				}()
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				log.Debugf("Rebalanced: %+v", ntf)
			}
		case err, more := <-consumer.Errors():
			if more {
				log.Debugf("Error: %s", err)
			}
		case <-sigchan:
			{
				log.Debugf("Exit signal!")
			}
			return
		}
	}
}
