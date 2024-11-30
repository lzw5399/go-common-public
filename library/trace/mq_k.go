package trace

import (
	"github.com/Shopify/sarama"
	"github.com/SkyAPM/go2sky/propagation"
)

func kExtractor(msg *sarama.ConsumerMessage) propagation.Extractor {
	return func() (s string, e error) {
		if msg == nil || msg.Headers == nil {
			return "", nil
		}
		for _, header := range msg.Headers {
			if string(header.Key) == propagation.Header {
				return string(header.Value), nil
			}
		}
		return "", nil
	}
}

func kInjector(msg *sarama.ProducerMessage) propagation.Injector {
	return func(header string) error {
		if msg == nil {
			return nil
		}
		if msg.Headers == nil {
			msg.Headers = []sarama.RecordHeader{}
		}
		sw8Header := sarama.RecordHeader{
			Key:   []byte(propagation.Header),
			Value: []byte(header),
		}
		msg.Headers = append(msg.Headers, sw8Header)
		return nil
	}
}
