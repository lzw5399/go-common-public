package fstorage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/pkg/errors"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var _ IStorage = new(diskStorage)

type diskStorage struct {
	cli *cos.Client
}

func newDiskStorage() (*diskStorage, error) {
	return &diskStorage{}, nil
}

func (s *diskStorage) Put(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	cfg := fconfig.DefaultConfig

	// 如果目标目录不存在，创建
	basePath := filepath.Dir(cfg.StorageNasDiskBasePath)
	if !exists(basePath) {
		err := os.MkdirAll(basePath, 0744)
		if err != nil {
			return nil, errors.Wrap(err, "diskStorage PutObject os.MkdirAll failed")
		}
	}
	// 创建目标文件
	dst := filepath.Join(basePath, objectName)
	out, err := os.Create(dst)
	if err != nil {
		return nil, errors.Wrap(err, "diskStorage PutObject os.Create failed")
	}
	defer out.Close()
	// 写入文件内容
	size, err := io.Copy(out, reader)

	rsp := &PutResult{
		Size:     size,
		Location: "",
	}

	return rsp, nil
}

func (s *diskStorage) FPut(ctx context.Context, objectName string, filePath string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	return s.Put(ctx, objectName, reader, objectSize, contentType)
}

func (s *diskStorage) Get(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	cfg := fconfig.DefaultConfig
	path := filepath.Join(cfg.StorageNasDiskBasePath, objectName)
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, 0, "", errors.Wrap(err, "diskStorage GetObject os.Stat failed")
	}

	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, "", errors.Wrap(err, "diskStorage GetObject os.Open failed")
	}

	return file, fileInfo.Size(), "", nil
}

func (s *diskStorage) Del(ctx context.Context, objectName string) error {
	cfg := fconfig.DefaultConfig
	path := filepath.Join(cfg.StorageNasDiskBasePath, objectName)
	err := os.Remove(path)
	if err != nil {
		return errors.Wrap(err, "diskStorage Del failed")
	}

	return nil
}

func (s *diskStorage) DeleteMulti(ctx context.Context, objectNames []string) error {
	cfg := fconfig.DefaultConfig
	for _, v := range objectNames {
		path := filepath.Join(cfg.StorageNasDiskBasePath, v)
		err := os.Remove(path)
		if err != nil {
			return errors.Wrap(err, "diskStorage Del failed")
		}
	}
	return nil
}

// exists reports whether the named file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
