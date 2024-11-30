package fredis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	goRedis "github.com/go-redis/redis/v8"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

const (
	SentinelClusterMode = "sentinelCluster"
	SentinelNormalMode  = "sentinelNormal"
)

var (
	gUniClient    goRedis.UniversalClient
	gSentinelMode string

	singleClient          *goRedis.Client
	sentinelClient        *goRedis.Client
	sentinelClusterClient *goRedis.ClusterClient // redis sentinel 模式，主节点写从节点读
	clusterClient         *goRedis.ClusterClient
)

var (
	addrLackErr = errors.New("client init addr empty")
)

// Init 用来初始化redis，mode--redis模式，可选项为: single--单例，sentinel--主从，cluster--集群
// opt是对应初始化的配置，需要根据mode采用相应的类型: single:go-redis.Options,sentinel:go-redis.FailoverOptions,cluster:go-redis.ClusterOptions
// cfg.RedisSentinelMode参数用来为主从模式指定是否采用"主节点写从节点读"的读写策略（默认采用）。不采用该策略时cfg.RedisSentinelMode参数填为: SentinelNormalMode
func Init() error {
	cfg := fconfig.DefaultConfig.RedisConfig
	var (
		err error
	)
	switch cfg.RedisMode {
	case fconfig.REDIS_MODE_SINGLE:
		opt := &goRedis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
		}
		err = initSingleClient(opt)
	case fconfig.REDIS_MODE_SENTINEL:
		opt := &goRedis.FailoverOptions{
			SentinelAddrs: strings.Split(cfg.RedisSentinelAddr, ","),
			Password:      cfg.RedisSentinelPassword,
			MasterName:    cfg.RedisMasterName,
		}
		gSentinelMode = cfg.RedisSentinelMode
		if gSentinelMode == SentinelNormalMode {
			err = initSentinelFailoverClient(opt)
			break
		}
		err = initSentinelFailoverClusterClient(opt)
	case fconfig.REDIS_MODE_CLUSTER:
		opt := &goRedis.ClusterOptions{
			Addrs:    strings.Split(cfg.RedisAddr, ","),
			Password: cfg.RedisPassword,
		}
		err = initClusterClient(opt)
	default:
		return errors.New("mode must be: single、sentinel or cluster")
	}

	// 忽略redis的内部打印的日志
	goRedis.SetLogger(&NothingLogAdaptor{})

	return err
}

func initSingleClient(opt *goRedis.Options) (err error) {
	fmt.Println("[fcredis] initSingleClient.")
	if opt.Addr == "" {
		return addrLackErr
	}
	singleClient = goRedis.NewClient(opt)
	gUniClient = singleClient
	return nil
}

func initSentinelFailoverClient(opt *goRedis.FailoverOptions) (err error) {
	fmt.Println("[fcredis] initSentinelFailoverClient.")
	if len(opt.SentinelAddrs) == 0 {
		return addrLackErr
	}
	sentinelClient = goRedis.NewFailoverClient(opt)
	_, err = sentinelClient.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	gUniClient = sentinelClient
	return nil
}

func initSentinelFailoverClusterClient(opt *goRedis.FailoverOptions) (err error) {
	fmt.Println("[fcredis] initSentinelFailoverClusterClient.")
	if len(opt.SentinelAddrs) == 0 {
		return addrLackErr
	}

	sentinelClusterClient = goRedis.NewFailoverClusterClient(opt)
	_, err = sentinelClusterClient.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	gUniClient = sentinelClusterClient
	return nil
}

func initClusterClient(opt *goRedis.ClusterOptions) (err error) {
	fmt.Println("[fcredis] initClusterClient.")
	if len(opt.Addrs) == 0 {
		return addrLackErr
	}

	clusterClient = goRedis.NewClusterClient(opt)
	_, err = clusterClient.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	gUniClient = clusterClient

	return nil
}
