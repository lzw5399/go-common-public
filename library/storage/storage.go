package fstorage

import (
	"context"
	"io"
	"sync"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/pkg/errors"
)

var (
	defaultStorage IStorage
	once           sync.Once
)

func InitStorage() {
	once.Do(func() {
		cfg := fconfig.DefaultConfig

		var err error
		switch cfg.StorageMode {
		case "aws_s3":
			defaultStorage, err = newAwsS3Storage()

		case "tencent_cos":
			defaultStorage, err = newTencentCosStorage()

		case "disk":
			defaultStorage, err = newDiskStorage()

		case "minio":
			defaultStorage, err = newMinioStorage()

		case "ali_oss":
			defaultStorage, err = newAliOssStorage()

		default:
			panic("InitStorage failed invalid storage mode")
		}

		if err != nil {
			panic(errors.Wrap(err, "InitStorage failed"))
		}
	})
}

func Put(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (rsp *PutResult, err error) {
	return defaultStorage.Put(ctx, objectName, reader, objectSize, contentType)
}

func FPut(ctx context.Context, objectName string, filePath string, reader io.Reader, objectSize int64, contentType string) (rsp *PutResult, err error) {
	return defaultStorage.FPut(ctx, objectName, filePath, reader, objectSize, contentType)
}

func Get(ctx context.Context, objectName string) (reader io.ReadCloser, objectSize int64, contentType string, err error) {
	return defaultStorage.Get(ctx, objectName)
}

func Del(ctx context.Context, objectName string) error {
	return defaultStorage.Del(ctx, objectName)
}

func DeleteMulti(ctx context.Context, objectNames []string) error {
	return defaultStorage.DeleteMulti(ctx, objectNames)
}

type IStorage interface {
	Put(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (rsp *PutResult, err error)
	FPut(ctx context.Context, objectName string, filePath string, reader io.Reader, objectSize int64, contentType string) (rsp *PutResult, err error)
	Get(ctx context.Context, objectName string) (reader io.ReadCloser, objectSize int64, contentType string, err error)
	Del(ctx context.Context, objectName string) error
	DeleteMulti(ctx context.Context, objectNames []string) error
}

type PutResult struct {
	Size     int64
	Location string
}
