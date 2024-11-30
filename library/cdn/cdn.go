package cdn

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/log"
)

var (
	cdnClient ICdn
	once      sync.Once
)

func InitCdn() {
	var err error
	once.Do(func() {
		cfg := fconfig.DefaultConfig
		switch strings.ToLower(cfg.CdnCachePushMode) {
		case "tencent":
			cdnClient, err = NewTencentClient()
		case "ali":
			cdnClient, err = NewAliClient()
		default:
			panic("Init cdn failed invalid cdn mode")
		}
	})
	if err != nil {
		panic(errors.Wrap(err, "Init cdn push failed"))
	}
}

// PushCdnCache 目前默认地区是国内cdn预热
func PushCdnCache(ctx context.Context, path []string) error {
	cfg := fconfig.DefaultConfig
	if cfg.CdnOpen && cfg.CdnCachePushOpen {
		log.Infof("start to push cdn cache, path is [%v]", path)
		return cdnClient.PushObjectCache(ctx, path)
	}
	return nil
}

// GetCdnPushCachePath 获取预热的path
func GetCdnPushCachePath(netdiskId string) string {
	if netdiskId == "" {
		return ""
	}

	if strings.Contains(netdiskId, "/") {
		netdiskId = GetNetdiskIdFromDownloadUrl(netdiskId)
	}

	cfg := fconfig.DefaultConfig
	return fmt.Sprintf("%s%s%s", cfg.CdnDomain, cfg.CdnUri, netdiskId)
}

// GetCdnDownloadUrl 获取cdn的下载链接
func GetCdnDownloadUrl(uri string) string {
	if uri == "" {
		return ""
	}

	// 本来就是http链接，直接返回
	if strings.HasPrefix(strings.ToLower(uri), "http") {
		return uri
	}

	cfg := fconfig.DefaultConfig
	if !cfg.CdnOpen {
		return uri
	}

	if cfg.CdnMode == fconfig.CDN_MODE_REDIRECT {
		return fmt.Sprintf("%s%s", cfg.CdnPreUri, GetNetdiskIdFromDownloadUrl(uri))
	}

	return fmt.Sprintf("%s%s", cfg.CdnDomain, uri)
}

// GetCdnDownloadDirectUrl 获取cdn的直连下载链接
func GetCdnDownloadDirectUrl(uri string) string {
	if uri == "" {
		return ""
	}

	cfg := fconfig.DefaultConfig
	if !cfg.CdnOpen {
		return uri
	}

	return fmt.Sprintf("%s%s", cfg.CdnDomain, uri)
}

// GetNetdiskIdFromDownloadUrl 从带路由的下载地址获取 netdiskId
func GetNetdiskIdFromDownloadUrl(fullUrl string) string {
	if fullUrl == "" {
		return ""
	}

	tempList := strings.Split(fullUrl, "/")
	idStr := tempList[len(tempList)-1]

	// 移除字符串中 ?之后的部分
	return strings.Split(idStr, "?")[0]
}

// RefreshCdnCache 清除缓存
func RefreshCdnCache(ctx context.Context, path []string) error {
	cfg := fconfig.DefaultConfig
	if cfg.CdnOpen && cfg.CdnCachePushOpen {
		log.Infof("start to refresh cdn cache, path is [%v]", path)
		return cdnClient.RefreshObjectCache(ctx, path)
	}
	return nil
}

type ICdn interface {
	PushObjectCache(ctx context.Context, objectPath []string) (err error)
	RefreshObjectCache(ctx context.Context, path []string) (err error)
}
