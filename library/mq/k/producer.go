package k

import (
	"context"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/log"
	"github.com/pkg/errors"
)

var producer sarama.SyncProducer
var once sync.Once

func initProducerK() {
	cfg := fconfig.DefaultConfig
	version, err := sarama.ParseKafkaVersion(cfg.KVersion)
	if err != nil {
		panic(errors.Wrap(err, "k ParseKafkaVersion failed"))
	}

	kConfig := sarama.NewConfig()
	// 等待服务器所有副本都保存成功后的响应
	kConfig.Producer.RequiredAcks = sarama.WaitForAll
	// 随机的分区类型：返回一个分区器，该分区器每次选择一个随机分区
	kConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	// 是否等待成功和失败后的响应
	kConfig.Producer.Return.Successes = true
	kConfig.Version = version

	if cfg.KUser != "" && cfg.KPwd != "" {
		kConfig.Net.SASL.Enable = true
		kConfig.Net.SASL.User = cfg.KUser
		kConfig.Net.SASL.Password = cfg.KPwd
		kConfig.Net.SASL.Handshake = true

		if cfg.KMechanism == sarama.SASLTypeSCRAMSHA256 {
			kConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
			kConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		} else if cfg.KMechanism == sarama.SASLTypeSCRAMSHA512 {
			kConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
			kConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		} else {
			kConfig.Net.SASL.Mechanism = sarama.SASLMechanism(cfg.KMechanism)
		}
	}

	// 使用给定代理地址和配置创建一个同步生产者
	producer, err = sarama.NewSyncProducer(strings.Split(cfg.KAddr, ","), kConfig)
	if err != nil {
		panic(err)
	}
}

func SendKMsg(ctx context.Context, topic, value string) error {
	once.Do(initProducerK)

	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	// 将字符串转换为字节数组
	msg.Value = sarama.ByteEncoder(value)
	_, _, err := producer.SendMessage(msg)

	if err != nil {
		log.Errorf("Send k[%s] message[%s], err:%s,", topic, value, err.Error())
		return err
	}

	log.Infof("Send k[%s] message success, msg=%s", topic, value)
	return nil
}
