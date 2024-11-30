package registry

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

const KV_PUBLIC_NAME = "public"

type RegistryConfManager struct {
	registryKVClient *api.KV
	config           interface{} // 调用方的config对象指针, 缓存供watch使用
	started          bool
	lastValues       sync.Map // 缓存最新的值
	path             string
	address          string
}

func (c *RegistryConfManager) IsKeyLower() bool {
	return false
}

func (c *RegistryConfManager) InitFromRegistry(config interface{}, addr, path string) error {
	c.address = addr
	c.path = path
	c.config = config
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path is empty")
	}
	client, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		return err
	}
	c.registryKVClient = client.KV()
	pair, _, err := c.registryKVClient.Get(path, nil)
	if err != nil {
		return err
	}
	if pair == nil {
		// 如果未找到, 初始化一个空json对象
		// 否则, 后面的监听会失败
		pair = &api.KVPair{Key: path, Value: []byte("{}")}
		_, err = c.registryKVClient.Put(pair, nil)
		if err != nil {
			fmt.Printf("[remote-conf] put err: %v\n", err)
			return err
		}
	}
	newValues := make(map[string]interface{})
	err = json.Unmarshal(pair.Value, &newValues)
	if err != nil {
		return err
	}
	for k, newValue := range newValues {
		c.lastValues.Store(k, newValue)
	}
	fmt.Printf("[remote-conf] path: %v ,allkeys: %v\n", path, c.GetAllKeys())
	err = applyRemoteConfigForJson(c.config, c)
	if err != nil {
		return err
	}
	return nil

}

func (c *RegistryConfManager) StartWatch() error {
	if c.path == "" {
		return fmt.Errorf("RegistryConfManager should init first!")
	}
	if c.started {
		return fmt.Errorf("RegistryConfManager started!")
	}
	c.started = true
	params := map[string]interface{}{"type": "key", "key": c.path}
	plan, err := watch.Parse(params)
	plan.Handler = func(idx uint64, raw interface{}) {
		var v *api.KVPair
		if raw == nil { // nil is a valid return value
			v = nil
		} else {
			var ok bool
			if v, ok = raw.(*api.KVPair); !ok {
				return // ignore
			}
			newValues := make(map[string]interface{})
			err = json.Unmarshal(v.Value, &newValues)
			// 删除的配置项忽略
			for k, newValue := range newValues {
				oldValue, ok := c.lastValues.Load(k)
				if ok && oldValue == newValue {
					continue
				}
				c.lastValues.Store(k, newValue)
				err = updateForJson(c.config, strings.ToUpper(k), fmt.Sprintf("%v", newValues[k]))
				if err != nil {
					fmt.Printf("[remote-conf] error: update config err: %v\n", err)
					continue
				} else {
					fmt.Printf("[remote-conf] update config success [%v]: %v -> %v\n", k, oldValue, newValue)
				}
			}
		}
	}
	go plan.Run(c.address)
	return nil
}

// Get 统一返回字符串, 后续单独处理
func (c *RegistryConfManager) Get(key string) string {
	if value, ok := c.lastValues.Load(key); ok {
		return fmt.Sprintf("%v", value)
	} else {
		return ""
	}
}

func (c *RegistryConfManager) GetAllKeys() []string {
	res := []string{}
	c.lastValues.Range(func(key, value interface{}) bool {
		res = append(res, fmt.Sprintf("%v", key))
		return true
	})
	return res
}

// StartRegistryConfig 供客户端使用的标准方法
// 1. 初始化公共配置
// 2. 初始化服务私有配置并监听
// config必须是一个指针
func StartRegistryConfig(config interface{}, kvPath, serverName, addr string) error {
	if strings.TrimSpace(addr) == "" {
		return fmt.Errorf("registry addr is empty!")
	}
	if strings.TrimSpace(kvPath) == "" {
		return fmt.Errorf("kvPath is empty!")
	}
	if strings.TrimSpace(serverName) == "" {
		return fmt.Errorf("serverName is empty!")
	}
	remoteAddr := addr
	publicKeyPath := fmt.Sprintf("%s/%s", kvPath, KV_PUBLIC_NAME)
	privateKeyPath := fmt.Sprintf("%s/%s", kvPath, serverName)
	// 优先级 服务配置 > 公共配置 > 环境变量 > 代码中的默认值
	{
		registryConf := &RegistryConfManager{}
		if err := registryConf.InitFromRegistry(config, remoteAddr, publicKeyPath); err != nil {
			fmt.Printf("read config from registry [%v] [%v] fail: %v\n", remoteAddr, publicKeyPath, err)
			return err
		}
	}
	{
		registryConf := &RegistryConfManager{}
		if err := registryConf.InitFromRegistry(config, remoteAddr, privateKeyPath); err != nil {
			fmt.Printf("read config from registry [%v] [%v] fail: %v\n", remoteAddr, privateKeyPath, err)
			return err
		}
		if err := registryConf.StartWatch(); err != nil {
			fmt.Printf("config watch fail [%v] [%v] fail: %v\n", remoteAddr, privateKeyPath, err)
			return err
		}
	}

	fmt.Printf("config watch success [%v] [%v] \n", remoteAddr, privateKeyPath)
	return nil
}
