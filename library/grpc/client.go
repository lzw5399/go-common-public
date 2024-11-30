package fgrpc

import (
	_ "github.com/mbobakov/grpc-consul-resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/grpc/interceptor"
	"github.com/lzw5399/go-common-public/library/util"
)

func GetConn(serverName string) (*grpc.ClientConn, error) {
	cfg := fconfig.DefaultConfig
	switch cfg.GRPCDiscoveryMode {
	case "direct": // serverName是「服务名或ip:port」
		return dialGrpcConn(serverName)
	default:
		return getGrpcConnManager().getConn(cfg.RegistryAddr, serverName, "tag="+cfg.RegistryTag)
	}
}

func dialGrpcConn(target string) (*grpc.ClientConn, error) {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.OutgoingMetadataInterceptor),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(util.MB * 500)),
	}
	if fconfig.DefaultConfig.GRPCRoundRobin {
		options = append(options, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	}

	conn, err := grpc.Dial(
		target,
		options...,
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
