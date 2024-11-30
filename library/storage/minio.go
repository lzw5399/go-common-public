package fstorage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

var _ IStorage = new(minioStorage)

type minioStorage struct {
	cli *minio.Client
}

func newMinioStorage() (*minioStorage, error) {
	ctx := context.Background()
	cfg := fconfig.DefaultConfig

	client, err := minio.New(cfg.StorageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.StorageAccessKey, cfg.StorageSecretKey, ""),
		Secure: cfg.StorageS3UseSSL,
	})
	if err != nil {
		panic(err)
	}

	// 初始化存储桶
	err = client.MakeBucket(ctx, cfg.StorageBucketName, minio.MakeBucketOptions{
		Region: cfg.StorageS3Region,
	})
	if err != nil {
		// 初始化失败的话，检查存储桶是否存在，存在则继续
		exists, err := client.BucketExists(ctx, cfg.StorageBucketName)
		if err != nil {
			return nil, errors.Wrap(err, "newMinioStorage BucketExists failed")
		} else if !exists {
			return nil, errors.New("newMinioStorage check bucket not exist")
		}
	}

	return &minioStorage{
		cli: client,
	}, nil
}

func (m *minioStorage) Put(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	cfg := fconfig.DefaultConfig
	result, err := m.cli.PutObject(ctx, cfg.StorageBucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
		PartSize:    5 * 1024 * 1024, // 5MB 进行分片
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("minioStorage PutObject failed, objectSize: %d, contentType: %s", objectSize, contentType))
	}

	rsp := &PutResult{
		Size:     result.Size,
		Location: "",
	}
	return rsp, nil
}

func (m *minioStorage) FPut(ctx context.Context, objectName string, filePath string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	cfg := fconfig.DefaultConfig
	result, err := m.cli.FPutObject(ctx, cfg.StorageBucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("minioStorage FPutObject failed, objectSize: %d, contentType: %s", objectSize, contentType))
	}

	rsp := &PutResult{
		Size:     result.Size,
		Location: "",
	}
	return rsp, nil
}

func (m *minioStorage) Get(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	cfg := fconfig.DefaultConfig
	obj, err := m.cli.GetObject(ctx, cfg.StorageBucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, "", errors.Wrap(err, "minioStorage GetObject failed")
	}

	info, err := obj.Stat()
	if err != nil {
		return nil, 0, "", errors.Wrap(err, "minioStorage Get obj.Stat failed")
	}

	return obj, info.Size, info.ContentType, nil
}

func (m *minioStorage) Del(ctx context.Context, objectName string) error {
	cfg := fconfig.DefaultConfig
	err := m.cli.RemoveObject(ctx, cfg.StorageBucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.Wrap(err, "minioStorage Del failed")
	}

	return nil
}

func (m *minioStorage) DeleteMulti(ctx context.Context, objectNames []string) error {

	objectsCh := make(chan minio.ObjectInfo, len(objectNames))

	for _, v := range objectNames {
		object := minio.ObjectInfo{
			Key: v,
		}
		objectsCh <- object
	}

	close(objectsCh)
	// 删除对象
	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	cfg := fconfig.DefaultConfig
	for rErr := range m.cli.RemoveObjects(ctx, cfg.StorageBucketName, objectsCh, opts) {
		if rErr.Err != nil {
			return errors.Wrap(rErr.Err, "minioStorage DeleteMulti failed")
		}
	}

	return nil
}
