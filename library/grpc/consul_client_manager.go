package fgrpc

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	defaultGrpcConnectManager grpcConnectManager
	initOnce                  sync.Once
)

func getGrpcConnManager() *grpcConnectManager {
	initOnce.Do(func() {
		defaultGrpcConnectManager.GrpcConnectItemMap = make(map[string]grpcConnectItem)
	})

	return &defaultGrpcConnectManager
}

type grpcConnectItem struct { // 单个连接建立，例如：和账号系统多实例连接上，一个item可以理解为对应一个mop服务多实例的连接
	ClientConn unsafe.Pointer
	ServerName string
}

type grpcConnectManager struct {
	CreateLock         sync.Mutex                 // 防止多协程同时建立连接
	GrpcConnectItemMap map[string]grpcConnectItem // 一开始是空
}

// GetConn param是指url的参数，例如：wait=10s&tag=mop，代表这个连接超时10秒建立，registry的服务tag是mop
func (g *grpcConnectManager) getConn(registryUrl, server, param string) (*grpc.ClientConn, error) {
	if connItem, ok := g.GrpcConnectItemMap[server]; ok { // first check
		if atomic.LoadPointer(&connItem.ClientConn) != nil {
			return (*grpc.ClientConn)(connItem.ClientConn), nil
		}
	}

	g.CreateLock.Lock()
	defer g.CreateLock.Unlock()

	fmt.Println("getConn lock")
	connItem, ok := g.GrpcConnectItemMap[server]
	if ok { // double check
		if atomic.LoadPointer(&connItem.ClientConn) != nil {
			cc := (*grpc.ClientConn)(connItem.ClientConn)
			if g.checkState(cc) == nil {
				return cc, nil
			} else { // 旧的连接存在服务端断开的情况，需要先关闭
				_ = cc.Close()
			}
		}
	}

	target := "consul://" + registryUrl + "/" + server
	if param != "" {
		target += "?" + param
	}

	fmt.Println("new conn, url=" + target)
	cli, err := dialGrpcConn(target)
	if err != nil {
		return nil, err
	}

	var newItem grpcConnectItem
	newItem.ServerName = server

	atomic.StorePointer(&newItem.ClientConn, unsafe.Pointer(cli))
	g.GrpcConnectItemMap[server] = newItem

	return cli, nil
}

func (g *grpcConnectManager) checkState(conn *grpc.ClientConn) error {
	state := conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return errors.New("ErrConn")
	}

	return nil
}
