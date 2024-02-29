package aws

import (
	"bytes"
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/sirupsen/logrus"
)

type Aws interface {
	UploadImage(multipart.File, *multipart.FileHeader) (string, error)
}

var (
	S3 Aws
)

type awsClient struct {
	s3     *manager.Uploader
	config config.Aws
	logger *logrus.Logger
}

func NewAwsS3Service() {
	log := logger.Logger
	cfg := config.Config.Aws
	awsConfig, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion("eu-west-2"),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		log.WithError(err).Errorf("new s3 client")
	} else {
		log.Infoln("Aws S3 credential...OK")
	}

	s3Client := manager.NewUploader(s3.NewFromConfig(awsConfig))
	S3 = &awsClient{s3Client, cfg, log}
}

func (a awsClient) UploadImage(
	file multipart.File,
	fileHeader *multipart.FileHeader,
) (string, error) {
	buffer := make([]byte, fileHeader.Size)
	file.Read(buffer)

	params := &s3.PutObjectInput{
		Bucket: aws.String(a.config.S3.Buckets.Media),
		Key:    aws.String(fileHeader.Filename),
		Body:   bytes.NewReader(buffer),
	}

	uploadRes, err := a.s3.Upload(context.Background(), params)
	if err != nil {
		a.logger.WithError(err).Errorf("s3 image upload")
		return "", err
	}

	return uploadRes.Location, nil
}
