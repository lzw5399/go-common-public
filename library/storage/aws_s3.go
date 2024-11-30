package fstorage

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/pkg/errors"
)

var _ IStorage = new(awsS3Storage)

type awsS3Storage struct {
	cli *s3.S3
}

func newAwsS3Storage() (*awsS3Storage, error) {
	cfg := fconfig.DefaultConfig

	// Configure to use S3 Server
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(cfg.StorageAccessKey, cfg.StorageSecretKey, ""),
		Endpoint:    aws.String(cfg.StorageEndpoint),
		Region:      aws.String(cfg.StorageS3Region), //aws.String("us-east-1"),
		//DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(cfg.StorageS3ForcePath), //进入看注释可知其作用
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, errors.Wrap(err, "newAwsS3Storage session.NewSession failed")
	}
	s3Client := s3.New(newSession)

	cparams := &s3.HeadBucketInput{
		Bucket: aws.String(cfg.StorageBucketName), // 必须
	}

	//选择桶
	_, err = s3Client.HeadBucket(cparams)
	if err != nil {
		fmt.Println("HeadBucket err:", err)
		if err.Error() == s3.ErrCodeNoSuchBucket {
			//调用CreateBucket创建一个新的存储桶。
			creatPara := &s3.CreateBucketInput{
				Bucket: aws.String(cfg.StorageBucketName), // 必须
			}
			_, err = s3Client.CreateBucket(creatPara)
			if err != nil {
				return nil, errors.Wrap(err, "newAwsS3Storage s3Client.CreateBucket failed")
			}
		}
		return nil, errors.Wrap(err, "newAwsS3Storage s3Client.HeadBucket failed")
	}

	return &awsS3Storage{
		cli: s3Client,
	}, nil
}

func (a *awsS3Storage) Put(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	typ := contentType
	if typ == "" {
		typ = "application/octet-stream"
	}

	cfg := fconfig.DefaultConfig
	myACL := aws.String(cfg.StorageS3ObjectAcl) //acl 设置
	upInput := &s3manager.UploadInput{
		Bucket:      aws.String(cfg.StorageBucketName),
		Key:         aws.String(objectName),
		Body:        reader,
		ContentType: aws.String(typ),
		ACL:         myACL, //object上的ACL
	}
	uploader := s3manager.NewUploaderWithClient(a.cli)
	result, err := uploader.Upload(upInput, func(u *s3manager.Uploader) {
		u.PartSize = 20 * 1024 * 1024 // 分块大小,当文件体积超过10M开始进行分块上传
		u.LeavePartsOnError = true
		u.Concurrency = 3
	}) //并发数
	if err != nil {
		return nil, errors.Wrap(err, "awsS3Storage Put failed")
	}

	rsp := &PutResult{
		Size:     objectSize,
		Location: result.Location,
	}
	return rsp, nil
}

func (a *awsS3Storage) FPut(ctx context.Context, objectName string, filePath string, reader io.Reader, objectSize int64, contentType string) (*PutResult, error) {
	return a.Put(ctx, objectName, reader, objectSize, contentType)
}

func (a *awsS3Storage) Get(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	cfg := fconfig.DefaultConfig
	getInputInfo := s3.GetObjectInput{
		Bucket: aws.String(cfg.StorageBucketName),
		Key:    aws.String(objectName),
	}
	fileInfo, err := a.cli.GetObject(&getInputInfo)
	if err != nil {
		return nil, 0, "", errors.Wrap(err, "awsS3Storage Get failed")
	}

	return fileInfo.Body, *fileInfo.ContentLength, *fileInfo.ContentType, nil
}

func (a *awsS3Storage) Del(ctx context.Context, objectName string) error {
	cfg := fconfig.DefaultConfig
	delInput := s3.DeleteObjectInput{
		Bucket: aws.String(cfg.StorageBucketName),
		Key:    aws.String(objectName),
	}
	_, err := a.cli.DeleteObject(&delInput)
	if err != nil {
		return errors.Wrap(err, "awsS3Storage Del failed")
	}

	return nil
}

func (a *awsS3Storage) DeleteMulti(ctx context.Context, objectNames []string) error {
	cfg := fconfig.DefaultConfig
	// 要删除的对象列表
	objects := []*s3.ObjectIdentifier{}
	for _, v := range objectNames {
		object := s3.ObjectIdentifier{}
		object.Key = aws.String(v)
		objects = append(objects, &object)
	}
	deleteInput := &s3.DeleteObjectsInput{
		Bucket: aws.String(cfg.StorageBucketName),
		Delete: &s3.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false), // Verbose模式
		},
	}
	_, err := a.cli.DeleteObjects(deleteInput)
	if err != nil {
		return errors.Wrap(err, "awsS3Storage DeleteMulti failed")
	}

	return nil
}
