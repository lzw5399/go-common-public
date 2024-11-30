package cdn

import (
	"context"
	"strings"

	ali "github.com/alibabacloud-go/cdn-20180510/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

var _ ICdn = new(aliClient)

type aliClient struct {
	client *ali.Client
}

func NewAliClient() (*aliClient, error) {
	cfg := fconfig.DefaultConfig
	aliConfig := &openapi.Config{
		AccessKeyId:     tea.String(cfg.CdnAccessKeyId),
		AccessKeySecret: tea.String(cfg.CdnAccessKeySecret),
	}
	client, err := ali.NewClient(aliConfig)
	if err != nil {
		return nil, err
	}
	return &aliClient{client: client}, nil
}

// PushObjectCache 需要预热的URL，多个URL之间需要用换行符\n或\r\n分隔。
func (a aliClient) PushObjectCache(ctx context.Context, path []string) (err error) {
	request := &ali.PushObjectCacheRequest{ObjectPath: tea.String(strings.Join(path, "\n"))}
	_, err = a.client.PushObjectCache(request)
	return
}

// RefreshObjectCache 清除缓存
func (a aliClient) RefreshObjectCache(ctx context.Context, path []string) (err error) {
	request := &ali.RefreshObjectCachesRequest{ObjectPath: tea.String(strings.Join(path, "\n"))}
	a.client.RefreshObjectCaches(request)
	return
}
