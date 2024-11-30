package fclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"google.golang.org/protobuf/proto"

	"github.com/lzw5399/go-common-public/library/util/encrypt"

	fredis "github.com/lzw5399/go-common-public/library/cache/redis"
	fconfig "github.com/lzw5399/go-common-public/library/config"
	ferrors "github.com/lzw5399/go-common-public/library/errors"
	restyutil "github.com/lzw5399/go-common-public/library/http/resty"
	"github.com/lzw5399/go-common-public/library/log"
	fpb "github.com/lzw5399/go-common-public/library/pb"
)

const (
	_CACHE_KEY_LICENSE_SERVICE_STR = "fc:license:service" // 业务服务使用的license
)

var _ LicenseManagerClient = new(licenseManagerClient)

type LicenseManagerClient interface {
	GetLicense(ctx context.Context) (*fpb.LicenseData, error)                                                 // 获取license
	CheckLicense(ctx context.Context, req *fpb.CheckLicenseReq) (*fpb.CheckLicenseRsp, error)                 // 校验即将更新的license
	UpdateLicense(ctx context.Context, req *fpb.UpdateLicenseReq) error                                       // 更新license
	GetLicenseContentByEncryptedStr(ctx context.Context, req *fpb.UpdateLicenseReq) (*fpb.LicenseData, error) // 获取即将更新的license中数据

	// ext license
	ListExtLicenseHistory(ctx context.Context) (*fpb.ListExtLicenseHistoryResp, error)
	UpdateExtLicense(ctx context.Context, req *fpb.UpdateExtLicenseReq) error
	CheckExtLicense(ctx context.Context, req *fpb.UpdateExtLicenseReq) (*fpb.CheckExtLicenseResp, error)
}

type licenseManagerClient struct {
	cli *resty.Client
}

func NewLicenseManagerClient() LicenseManagerClient {
	return &licenseManagerClient{
		cli: restyutil.New().SetBaseURL(fconfig.DefaultConfig.LicenseManagerAddr),
	}
}

func (c *licenseManagerClient) GetLicense(ctx context.Context) (*fpb.LicenseData, error) {
	licenseObj, rspInfo := fredis.GetOrSet(ctx, _CACHE_KEY_LICENSE_SERVICE_STR, time.Hour,
		func(ctx context.Context) ([]byte, *ferrors.SvrRspInfo) {
			licenseRsp, err := c.getLicense(ctx)
			if err != nil {
				log.Errorc(ctx, "licenseManagerCli.GetLicense err:%s", err)
				return nil, ferrors.InternalServerError()
			}
			raw, err := proto.Marshal(licenseRsp)
			if err != nil {
				log.Errorc(ctx, "licenseManagerClient proto marshal failed: %s", err)
				return nil, ferrors.InternalServerError()
			}

			return raw, ferrors.Ok()
		})
	if !rspInfo.Valid() {
		return nil, rspInfo
	}

	var license fpb.LicenseData
	err := proto.Unmarshal(licenseObj, &license)
	if err != nil {
		log.Errorc(ctx, "LicenseData proto.Unmarshal err:%s", err)
		return nil, err
	}

	if license.OrganName == "" {
		return nil, fmt.Errorf("LicenseData license is empty")
	}

	return &license, nil
}

func (c *licenseManagerClient) getLicense(ctx context.Context) (*fpb.LicenseData, error) {
	var resultWrapper LicenseRspWrapper
	rsp, err := c.cli.R().SetContext(ctx).
		SetHeaders(getHeader()).
		SetResult(&resultWrapper).
		Get("/api/v1/license")
	if err != nil {
		return nil, err
	}

	if !rsp.IsSuccess() {
		return nil, fmt.Errorf("GetLicense status code err, code:%d err:%s", rsp.StatusCode(), rsp.String())
	}

	if resultWrapper.Errcode != "OK" {
		return nil, fmt.Errorf("GetLicense err, errcode:%s err:%s", resultWrapper.Errcode, resultWrapper.Error)
	}

	decryptedStr := encrypt.CommonDecrypt(resultWrapper.Data.LicenseEncryptStr)

	var license fpb.LicenseData
	err = json.Unmarshal([]byte(decryptedStr), &license)
	if err != nil {
		return nil, fmt.Errorf("GetLicense unmarshal failed: %s", err)
	}

	// 反序列出来是空的
	if license.OrganName == "" {
		return nil, fmt.Errorf("")
	}

	return &license, nil
}

func (c *licenseManagerClient) CheckLicense(ctx context.Context, req *fpb.CheckLicenseReq) (*fpb.CheckLicenseRsp, error) {
	var resultWrapper CheckLicenseRspWrapper
	rsp, err := c.cli.R().SetContext(ctx).
		SetHeaders(getHeader()).
		SetBody(req).
		SetResult(&resultWrapper).
		Post("/api/v1/license/check")
	if err != nil {
		return nil, err
	}

	if !rsp.IsSuccess() {
		return nil, fmt.Errorf("CheckLicense status code err, code:%d err:%s", rsp.StatusCode(), rsp.String())
	}

	if resultWrapper.Errcode != "OK" {
		return nil, fmt.Errorf("CheckLicense err, errcode:%s err:%s", resultWrapper.Errcode, resultWrapper.Error)
	}

	return resultWrapper.Data, nil
}

func (c *licenseManagerClient) UpdateLicense(ctx context.Context, req *fpb.UpdateLicenseReq) error {
	rsp, err := c.cli.R().SetContext(ctx).
		SetHeaders(getHeader()).
		SetBody(req).
		Post("/api/v1/license/update")
	if err != nil {
		return err
	}

	if !rsp.IsSuccess() {
		return fmt.Errorf("UpdateLicense status code err, code:%d err:%s", rsp.StatusCode(), rsp.String())
	}

	return nil
}

func (c *licenseManagerClient) GetLicenseContentByEncryptedStr(ctx context.Context, req *fpb.UpdateLicenseReq) (*fpb.LicenseData, error) {
	var resultWrapper LicenseRspWrapper
	rsp, err := c.cli.R().SetContext(ctx).
		SetHeaders(getHeader()).
		SetBody(req).
		SetResult(&resultWrapper).
		Post("/api/v1/license/update/get/data")
	if err != nil {
		return nil, err
	}

	if !rsp.IsSuccess() {
		return nil, fmt.Errorf("GetLicenseContentByEncryptedStr status code err, code:%d err:%s", rsp.StatusCode(), rsp.String())
	}

	if resultWrapper.Errcode != "OK" {
		return nil, fmt.Errorf("GetLicenseContentByEncryptedStr err, errcode:%s err:%s", resultWrapper.Errcode, resultWrapper.Error)
	}

	decryptedStr := encrypt.CommonDecrypt(resultWrapper.Data.LicenseEncryptStr)

	var license fpb.LicenseData
	err = json.Unmarshal([]byte(decryptedStr), &license)
	if err != nil {
		return nil, fmt.Errorf("GetLicenseContentByEncryptedStr unmarshal failed: %s", err)
	}

	return &license, nil
}

func (c *licenseManagerClient) ListExtLicenseHistory(ctx context.Context) (*fpb.ListExtLicenseHistoryResp, error) {
	var resultWrapper ListExtLicenseHistoryWrapper
	rsp, err := c.cli.R().SetContext(ctx).
		SetHeaders(getHeader()).
		SetResult(&resultWrapper).
		Get("/api/v1/license/ext/history/list")
	if err != nil {
		return nil, err
	}

	if !rsp.IsSuccess() {
		// 如果能解析成wrapper，则返回ferrors错误
		var wrapper WrapperBase
		_ = json.Unmarshal(rsp.Body(), &wrapper)
		if wrapper.Errcode != "" {
			return nil, ferrors.New(rsp.StatusCode(), ferrors.ErrorCode(wrapper.Errcode))
		}

		// 否则返回status code错误
		return nil, fmt.Errorf("ListExtLicenseHistory status code err, code:%d err:%s", rsp.StatusCode(), rsp.String())
	}

	if resultWrapper.Errcode != "OK" {
		return nil, ferrors.New(rsp.StatusCode(), ferrors.ErrorCode(resultWrapper.Errcode))
	}

	return resultWrapper.Data, nil
}

func (c *licenseManagerClient) UpdateExtLicense(ctx context.Context, req *fpb.UpdateExtLicenseReq) error {
	var resultWrapper struct {
		Errcode string `json:"errcode"`
		Error   string `json:"error"`
	}

	rsp, err := c.cli.R().SetContext(ctx).
		SetHeaders(getHeader()).
		SetBody(req).
		SetResult(&resultWrapper).
		Post("/api/v1/license/ext/update")
	if err != nil {
		return err
	}

	if !rsp.IsSuccess() {
		// 如果能解析成wrapper，则返回ferrors错误
		var wrapper WrapperBase
		_ = json.Unmarshal(rsp.Body(), &wrapper)
		if wrapper.Errcode != "" {
			return ferrors.New(rsp.StatusCode(), ferrors.ErrorCode(wrapper.Errcode))
		}

		// 否则返回status code错误
		return fmt.Errorf("UpdateExtLicense status code err, code:%d err:%s", rsp.StatusCode(), rsp.String())
	}

	if resultWrapper.Errcode != "OK" {
		return ferrors.New(rsp.StatusCode(), ferrors.ErrorCode(resultWrapper.Errcode))
	}

	return nil
}

func (c *licenseManagerClient) CheckExtLicense(ctx context.Context, req *fpb.UpdateExtLicenseReq) (*fpb.CheckExtLicenseResp, error) {
	var resultWrapper CheckExtLicenseWrapper
	rsp, err := c.cli.R().SetContext(ctx).
		SetHeaders(getHeader()).
		SetBody(req).
		SetResult(&resultWrapper).
		Post("/api/v1/license/ext/check")
	if err != nil {
		return nil, err
	}

	if !rsp.IsSuccess() {
		// 如果能解析成wrapper，则返回ferrors错误
		var wrapper WrapperBase
		_ = json.Unmarshal(rsp.Body(), &wrapper)
		if wrapper.Errcode != "" {
			return nil, ferrors.New(rsp.StatusCode(), ferrors.ErrorCode(wrapper.Errcode))
		}

		// 否则返回status code错误
		return nil, fmt.Errorf("CheckExtLicense status code err, code:%d err:%s", rsp.StatusCode(), rsp.String())
	}

	if resultWrapper.Errcode != "OK" {
		return nil, ferrors.New(rsp.StatusCode(), ferrors.ErrorCode(resultWrapper.Errcode))
	}

	return resultWrapper.Data, nil
}

type WrapperBase struct {
	Error   string `json:"error"`
	Errcode string `json:"errcode"`
}

type LicenseRspWrapper struct {
	Error   string                  `json:"error"`
	Errcode string                  `json:"errcode"`
	Data    *fpb.LicenseEncryptResp `json:"data"`
}

type ListExtLicenseHistoryWrapper struct {
	Error   string                         `json:"error"`
	Errcode string                         `json:"errcode"`
	Data    *fpb.ListExtLicenseHistoryResp `json:"data"`
}

type CheckExtLicenseWrapper struct {
	Error   string                   `json:"error"`
	Errcode string                   `json:"errcode"`
	Data    *fpb.CheckExtLicenseResp `json:"data"`
}

type CheckLicenseRspWrapper struct {
	Error   string               `json:"error"`
	Errcode string               `json:"errcode"`
	Data    *fpb.CheckLicenseRsp `json:"data"`
}

type CheckLicenseNumRspWrapper struct {
	Error   string                  `json:"error"`
	Errcode string                  `json:"errcode"`
	Data    *fpb.CheckLicenseNumRsp `json:"data"`
}

type LicenseExpireTimeRspWrapper struct {
	Error   string                             `json:"error"`
	Errcode string                             `json:"errcode"`
	Data    *fpb.GetUpdateLicenseExpireTimeRsp `json:"data"`
}

var (
	headers = map[string]string{
		"Accept":       "application/json, text/plain, */*",
		"Content-Type": "application/json",
		"f-caller":     "finclip-cloud",
	}
)

// getHeader map赋值是引用模式(不能重入)，所以建议使用这个函数获取header
func getHeader() map[string]string {
	newMap := make(map[string]string, len(headers))
	for k, v := range headers {
		newMap[k] = v
	}

	return newMap
}
