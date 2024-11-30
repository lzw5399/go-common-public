package trace

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/SkyAPM/go2sky"
)

func (client *Client) initK(config KConfig) {
	if config.extractor == nil {
		client.kExtractor = kExtractor
	} else {
		client.kExtractor = config.extractor
	}
	if config.injector == nil {
		client.kInjector = kInjector
	} else {
		client.kInjector = config.injector
	}
}

func (client *Client) CreateKEntrySpan(ctx context.Context, topic, method string, msg *sarama.ConsumerMessage) (go2sky.Span, context.Context) {
	span, nCtx := client.CreateEntrySpan(ctx, topic, client.kExtractor(msg))
	span.Tag(TagMQType, MQTypeK)
	span.Tag(go2sky.TagMQTopic, topic)
	span.Tag(TagMQMethod, method)
	return span, nCtx
}

func (client *Client) CreateKExitSpan(ctx context.Context, topic, method, address string, msg *sarama.ProducerMessage) go2sky.Span {
	span := client.CreateExitSpan(ctx, topic, address, client.kInjector(msg))
	span.Tag(TagMQType, MQTypeK)
	span.Tag(go2sky.TagMQTopic, topic)
	span.Tag(TagMQMethod, method)
	return span
}
