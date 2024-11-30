package fgrpc

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/discovery/registry"
	"github.com/lzw5399/go-common-public/library/grpc/interceptor"
	"github.com/lzw5399/go-common-public/library/util"
)

// StartGRPC 启动grpc服务
func StartGRPC(registerFunc func(s *grpc.Server)) {
	// 如果服务发现模式是registry的话，把服务注册到consul
	cfg := fconfig.DefaultConfig
	if cfg.GRPCDiscoveryMode == "registry" && cfg.RegistryAddr != "" {
		registry.RegisterService()
	}

	// 开启grpc服务
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(util.MB*500),
		grpc.MaxSendMsgSize(util.MB*500),
		grpc.ChainUnaryInterceptor(
			interceptor.InComingMetadataInterceptor,
			interceptor.ValidatorInterceptor(),
		),
	)
	registerFunc(s)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}
