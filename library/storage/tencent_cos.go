package fstorage

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/pkg/errors"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var _ IStorage = new(tencentCosStorage)

type tencentCosStorage struct {
	cli *cos.Client
}

func newTencentCosStorage() (*tencentCosStorage, error) {
	cfg := fconfig.DefaultConfig
	u, err := url.Parse(cfg.StorageEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "newTencentCosStorage parse endpoint failed")
	}

	b := &cos.BaseURL{
		BucketURL: u,
	}
	cosClient := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.StorageAccessKey,
			SecretKey: cfg.StorageSecretKey,
		},
	})

	return &tencentCosStorage{
		cli: cosClient,
	}, nil
}

func (s *tencentCosStorage) Put(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	typ := contentType
	if typ == "" {
		typ = "application/octet-stream"
	}

	putOpt := &cos.ObjectPutOptions{
		ACLHeaderOptions:       new(cos.ACLHeaderOptions),
		ObjectPutHeaderOptions: new(cos.ObjectPutHeaderOptions),
	}
	putOpt.ContentType = typ
	putOpt.ContentLength = objectSize
	putOpt.XCosACL = "public-read"
	result, err := s.cli.Object.Put(ctx, objectName, reader, putOpt)
	if err != nil {
		return nil, errors.Wrap(err, "tencentCosStorage PutObject failed")
	}

	rsp := &PutResult{
		Size: result.ContentLength,
	}

	location, err := result.Location()
	if err != nil {
		rsp.Location = s.cli.BaseURL.BucketURL.String() + objectName
	} else {
		rsp.Location = location.Path
	}

	cfg := fconfig.DefaultConfig
	if cfg.StorageUploadPath != "" && strings.Contains(rsp.Location, cfg.StorageUploadPath) {
		rsp.Location = strings.Replace(rsp.Location, cfg.StorageUploadPath, "/"+cfg.StorageUploadPath, -1)
	}

	return rsp, nil
}

func (s *tencentCosStorage) FPut(ctx context.Context, objectName string, filePath string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	return s.Put(ctx, objectName, reader, objectSize, contentType)
}

func (s *tencentCosStorage) Get(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	resp, err := s.cli.Object.Get(ctx, objectName, nil)
	if err != nil {
		return nil, 0, "", errors.Wrap(err, "tencentCosStorage GetObject failed")
	}

	return resp.Body, resp.ContentLength, "", nil
}

func (s *tencentCosStorage) Del(ctx context.Context, objectName string) error {
	_, err := s.cli.Object.Delete(ctx, objectName)
	if err != nil {
		return errors.Wrap(err, "tencentCosStorage Del failed")
	}

	return nil
}

func (s *tencentCosStorage) DeleteMulti(ctx context.Context, objectNames []string) error {
	// 创建删除选项
	deleteOpt := &cos.ObjectDeleteMultiOptions{
		Objects: make([]cos.Object, len(objectNames)),
		Quiet:   false, // Verbose模式
	}

	for i, obj := range objectNames {
		deleteOpt.Objects[i] = cos.Object{Key: obj}
	}
	_, _, err := s.cli.Object.DeleteMulti(ctx, deleteOpt)
	if err != nil {
		return errors.Wrap(err, "tencentCosStorage Del failed")
	}

	return nil
}
