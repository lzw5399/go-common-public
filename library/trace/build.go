package trace

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
)

var (
	lock    = &sync.Mutex{}
	client  *Client
	version = "v0.2.3"
)

type HttpConfig struct {
	extractor func(req *http.Request) propagation.Extractor
	injector  func(req *http.Request) propagation.Injector
}

type GRpcConfig struct {
	extractor func(ctx context.Context) propagation.Extractor
	injector  func(ctx *context.Context) propagation.Injector
}

type KConfig struct {
	extractor func(msg *sarama.ConsumerMessage) propagation.Extractor
	injector  func(msg *sarama.ProducerMessage) propagation.Injector
}

type MongoConfig struct {
}

type ESConfig struct {
}

type RedisConfig struct {
}

type MySqlConfig struct {
}

type Build struct {
	url              string
	serverName       string
	enable           bool
	samplePartitions uint32
	reporter         go2sky.Reporter
	httpConfig       HttpConfig
	gRpcConfig       GRpcConfig
	kConfig          KConfig
	mongoConfig      MongoConfig
	eSConfig         ESConfig
	redisConfig      RedisConfig
	mySqlConfig      MySqlConfig
}

func CreateBuild(url string, serverName string, samplePartitions uint32, enable bool) *Build {
	return &Build{
		url:              url,
		serverName:       serverName,
		samplePartitions: samplePartitions,
		enable:           enable,
		reporter:         nil,
		httpConfig:       HttpConfig{},
		gRpcConfig:       GRpcConfig{},
		kConfig:          KConfig{},
		mongoConfig:      MongoConfig{},
		eSConfig:         ESConfig{},
		redisConfig:      RedisConfig{},
		mySqlConfig:      MySqlConfig{},
	}
}

func (build *Build) buildClient() *Client {
	trace, reporter := createTracer(build.url, build.serverName,
		build.samplePartitions, build.enable, build.reporter)

	client := Client{
		tracer:   trace,
		reporter: reporter,
		enable:   build.enable,
	}
	client.initHttp(build.httpConfig)
	client.initGRpc(build.gRpcConfig)
	client.initK(build.kConfig)
	client.initMongo(build.mongoConfig)
	client.initES(build.eSConfig)
	client.initRedis(build.redisConfig)
	client.initMySql(build.mySqlConfig)

	return &client
}

func (build *Build) setKConfig(kConfig KConfig) {
	build.kConfig = kConfig
}

func BuildApmClient(build *Build) {
	lock.Lock()
	defer lock.Unlock()
	client = build.buildClient()
	log.Printf("BuildApmClient version[%s]", version)
}

func ApmClient() *Client {
	if client == nil {
		panic(errors.New("please initialize tracer before using it"))
	}
	return client
}

func CloseApmClient() {
	lock.Lock()
	defer lock.Unlock()

	if client != nil {
		client.Close()
		client = nil
	}
}
