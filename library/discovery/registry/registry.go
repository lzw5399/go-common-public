package registry

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	registryapi "github.com/hashicorp/consul/api"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

var (
	defaultDiscoveryManager *Registry
	once                    sync.Once
)

// Registry 结构体定义了服务注册所需的基本信息
type Registry struct {
	client    *registryapi.Client
	serviceID string
	checkPort string
	isStop    bool
}

// RegisterService 初始化 Registry，包括创建客户端、健康检查和注册服务
func RegisterService() {
	once.Do(func() {
		defaultDiscoveryManager = &Registry{
			serviceID: uuid.New().String(),
			checkPort: "9091",
		}

		cfg := registryapi.DefaultConfig()
		cfg.Address = fconfig.DefaultConfig.RegistryAddr
		client, err := registryapi.NewClient(cfg)
		if err != nil {
			panic(err)
		}
		defaultDiscoveryManager.client = client

		defaultDiscoveryManager.exportHealthCheckEndpoint() // 暴露健康检查的 HTTP 处理器
		defaultDiscoveryManager.registerService()           // 向注册中心注册服务
		go defaultDiscoveryManager.runReRegisterJob()       // 定期检查，如果服务不健康则重新注册
	})
}

// DeregisterService 从注册中心注销服务
func DeregisterService() {
	if defaultDiscoveryManager == nil {
		return
	}

	defaultDiscoveryManager.isStop = true
	if err := defaultDiscoveryManager.client.Agent().ServiceDeregister(defaultDiscoveryManager.serviceID); err != nil {
		fmt.Printf("[go-common Registry] Failed to deregister service(%s), err: %s\n", fconfig.DefaultConfig.ServerName, err)
	} else {
		fmt.Printf("[go-common Registry] Service deregistered(%s) successfully: %s\n", fconfig.DefaultConfig.ServerName, defaultDiscoveryManager.serviceID)
	}
}

// exportHealthCheckEndpoint 初始化健康检查的 HTTP 处理器
func (r *Registry) exportHealthCheckEndpoint() {
	http.HandleFunc("/"+r.serviceID+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(":"+r.checkPort, nil)
}

// getLocalIP 获取本地 IP 地址
func (r *Registry) getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// registerService 向注册中心注册服务
func (r *Registry) registerService() {
	registration := &registryapi.AgentServiceRegistration{
		ID:      r.serviceID,
		Name:    fconfig.DefaultConfig.ServerName,
		Port:    mustAtoi(fconfig.DefaultConfig.GRPCPort),
		Tags:    []string{fconfig.DefaultConfig.RegistryTag},
		Address: r.getLocalIP(),
		Check: &registryapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%s/%s/health", r.getLocalIP(), r.checkPort, r.serviceID),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		panic(fmt.Sprintf("[go-common Registry] Failed to register service(%s), err:%s", fconfig.DefaultConfig.ServerName, err))
	}
	fmt.Printf("[go-common Registry] Service registered (%s) successfully: %s\n", fconfig.DefaultConfig.ServerName, r.serviceID)
}

// runReRegisterJob 如果当前服务不健康，重新注册服务
func (r *Registry) runReRegisterJob() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		if r.isStop {
			return
		}

		// 定时检查服务健康状态
		<-ticker.C

		// 服务不健康，重新注册服务
		if !r.checkServiceHealth() {
			r.registerService()
		}
	}
}

// checkServiceHealth 检查服务在注册中心的健康状态
func (r *Registry) checkServiceHealth() bool {
	services, _, err := r.client.Health().Service(fconfig.DefaultConfig.ServerName, fconfig.DefaultConfig.RegistryTag, true, nil)
	if err != nil { // 整个服务在consul中不存在
		fmt.Printf("[go-common Registry] Error checking service(%s) health: %s\n", fconfig.DefaultConfig.ServerName, err)
		return false
	}

	// 遍历所有服务，如果发现当前实例的服务 ID 不存在，则重新注册服务
	for _, service := range services {
		if service.Service.ID == r.serviceID {
			return true
		}
	}

	return false
}

// mustAtoi 将字符串转换为整数，如果转换失败则 panic
func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("[go-common Registry] mustAtoi failed to convert string to int: %v", err))
	}
	return i
}
