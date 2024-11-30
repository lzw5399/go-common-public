package cdn

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	tc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/log"
)

const EDN_POINT = "cdn.tencentcloudapi.com"

var _ ICdn = new(tencentClient)

type tencentClient struct {
	client *tc.Client
}

func NewTencentClient() (*tencentClient, error) {
	cfg := fconfig.DefaultConfig
	credential := common.NewCredential(
		cfg.CdnAccessKeyId,
		cfg.CdnAccessKeySecret,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = EDN_POINT
	client, err := tc.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}
	return &tencentClient{client: client}, nil
}

// PushObjectCache 需要预热的URL，多个URL之间需要用换行符\n或
func (a tencentClient) PushObjectCache(ctx context.Context, path []string) (err error) {

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := tc.NewPushUrlsCacheRequest()
	request.Urls = tea.StringSlice(path)

	// 返回的resp是一个PushUrlsCacheResponse的实例，与请求对象对应
	response, err := a.client.PushUrlsCache(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Errorf("tencent cdn push object cache error: %v", err)
		return
	}
	if err != nil {
		log.Errorf("tencent cdn push object cache error: %v", err)
		return err
	}
	log.Debugf("tencent cdn push object cache, resp is %s", response.ToJsonString())
	return err
}

func (a tencentClient) RefreshObjectCache(ctx context.Context, path []string) (err error) {

	// 创建请求
	request := tc.NewPurgePathCacheRequest()
	request.Paths = tea.StringSlice(path)
	request.FlushType = common.StringPtr("flush")

	// 发送请求
	response, err := a.client.PurgePathCache(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Errorf("tencent cdn RefreshObjectCache cache error: %v", err)

		return
	}
	log.Debugf("tencent cdn RefreshObjectCache cache, resp is %s", response.ToJsonString())
	return nil
}
