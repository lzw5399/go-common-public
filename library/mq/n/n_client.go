package n

import (
	"context"
	"time"

	"github.com/lzw5399/go-common-public/library/log"
	"github.com/nats-io/nats.go"
)

type ConsumerCb func(ctx context.Context, input []byte) error

type NClient struct {
	url  string
	conn *nats.Conn
}

func NewNClient(url string) *NClient {
	return &NClient{url: url}
}

func (s *NClient) reconnectCb(conn *nats.Conn) {
	log.Warnf("NClient reconnectCb triggered")
}

func (s *NClient) closedCb(conn *nats.Conn) {
	log.Warnf("NClient closedCb triggered")
}

func (s *NClient) disconnectErrCb(conn *nats.Conn, err error) {
	log.Warnf("NClient disconnect triggered, err:%s\n", err.Error())
}

func (s *NClient) discoveredServersCb(conn *nats.Conn) {
	log.Warnf("NClient discoveredServersCb triggered")
}

func (s *NClient) errorCb(conn *nats.Conn, sub *nats.Subscription, err error) {
	log.Warnf("NClient errorCb triggered err:%s\n", err.Error())
}

func (s *NClient) Start() {
	// 添加重连参数为-1，会无限重连
	conn, err := nats.Connect(s.url, nats.MaxReconnects(-1))
	if err != nil {
		log.Fatalf("NClient start url:%s err:%s\n", s.url, err.Error())
		return
	}
	log.Warnf("NClient start ok")
	s.conn = conn
	s.conn.SetReconnectHandler(s.reconnectCb)
	s.conn.SetClosedHandler(s.closedCb)
	s.conn.SetDisconnectErrHandler(s.disconnectErrCb)
	s.conn.SetDiscoveredServersHandler(s.discoveredServersCb)
	s.conn.SetErrorHandler(s.errorCb)
}

func (s *NClient) Pub(ctx context.Context, topic string, bytes []byte) error {
	err := s.conn.Publish(topic, bytes)
	if err != nil {
		log.Errorf("NClient pub msg to topic:%s error:%s\n", topic, err)
	}
	return err
}

func (s *NClient) Sub(ctx context.Context, topic string, cb ConsumerCb) {
	s.conn.Subscribe(topic, func(msg *nats.Msg) {
		cb(ctx, msg.Data)
	})
}

func (s *NClient) QueueSub(ctx context.Context, topic, queue string, cb ConsumerCb) {
	s.conn.QueueSubscribe(topic, queue, func(msg *nats.Msg) {
		cb(ctx, msg.Data)
	})
}

func (s *NClient) Request(topic string, bytes []byte, timeout int) ([]byte, error) {
	msg, err := s.conn.Request(topic, bytes, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		log.Errorf("NClient request topic:%s error:%s\n", topic, err)
		return nil, err
	}
	return msg.Data, nil
}

func (s *NClient) Response(msg *nats.Msg, data []byte) {
	s.conn.Publish(msg.Reply, data)
}
