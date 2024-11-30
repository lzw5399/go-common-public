package fstorage

import (
	"context"
	"io"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

var _ IStorage = new(aliOssStorage)

type aliOssStorage struct {
	bucket *oss.Bucket
}

func newAliOssStorage() (*aliOssStorage, error) {
	cfg := fconfig.DefaultConfig
	verifySsl := oss.InsecureSkipVerify(!cfg.StorageS3UseSSL)
	aliOssClient, err := oss.New(cfg.StorageEndpoint, cfg.StorageAccessKey, cfg.StorageSecretKey, verifySsl)
	if err != nil {
		return nil, errors.Wrap(err, "newAliOssStorage oss.New failed")
	}

	bk, err := aliOssClient.Bucket(cfg.StorageBucketName)
	if err != nil {
		return nil, errors.Wrap(err, "newAliOssStorage aliOssClient.Bucket failed")
	}

	return &aliOssStorage{
		bucket: bk,
	}, nil
}

func (o *aliOssStorage) Put(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ossOpt := []oss.Option{oss.ContentLength(objectSize), oss.ContentType(contentType)}
	err := o.bucket.PutObject(objectName, reader, ossOpt...)
	if err != nil {
		return nil, errors.Wrap(err, "aliOssStorage Put failed")
	}

	rsp := &PutResult{
		Size:     objectSize,
		Location: "",
	}
	return rsp, nil
}

func (o *aliOssStorage) FPut(ctx context.Context, objectName string, filePath string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	return o.Put(ctx, objectName, reader, objectSize, contentType)
}

func (o *aliOssStorage) Get(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	h, err := o.bucket.GetObjectDetailedMeta(objectName)
	if err != nil || h == nil {
		return nil, 0, "", errors.Wrap(err, "aliOssStorage Get failed")
	}

	contentType := h.Get("Content-type")
	var objectSize int64

	size := h.Get("Content-Length")
	if size != "" {
		objectSize, _ = strconv.ParseInt(size, 10, 64)
	}

	reader, err := o.bucket.GetObject(objectName)
	if err != nil {
		return nil, 0, "", errors.Wrap(err, "aliOssStorage GetObject failed")
	}

	return reader, objectSize, contentType, nil
}

func (o *aliOssStorage) Del(ctx context.Context, objectName string) error {
	err := o.bucket.DeleteObject(objectName)
	if err != nil {
		return errors.Wrap(err, "aliOssStorage Del failed")
	}

	return nil
}

func (o *aliOssStorage) DeleteMulti(ctx context.Context, objectNames []string) error {
	_, err := o.bucket.DeleteObjects(objectNames)
	if err != nil {
		return errors.Wrap(err, "aliOssStorage Del failed")
	}
	return nil
}
