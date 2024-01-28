package aws

import (
	"bytes"
	"context"
	"mime/multipart"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

var awsService Aws

type awsClient struct {
	s3     *manager.Uploader
	config config.Aws
	logger *logrus.Logger
}

func GetS3Service() Aws { return awsService }

func NewAwsS3Service(config config.Aws, logger *logrus.Logger) Aws {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("eu-west-2"), awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretAccessKey, "")))
	if err != nil {
		panic(err)
	} else if err == nil {
		logger.Infoln("Aws S3 credential...OK")
	}

	s3Client := manager.NewUploader(s3.NewFromConfig(cfg))
	awsService = &awsClient{s3Client, config, logger}

	return awsService
}

func (a awsClient) UploadImage(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	buffer := make([]byte, fileHeader.Size)
	file.Read(buffer)

	params := &s3.PutObjectInput{
		Bucket: aws.String(a.config.S3.Buckets.Media),
		Key:    aws.String(fileHeader.Filename),
		Body:   bytes.NewReader(buffer),
	}

	uploadRes, err := a.s3.Upload(context.Background(), params)
	if err != nil {
		uploadErr := model.UziErr{Err: err.Error(), Message: "s3imageupload", Code: 400}
		a.logger.Errorf(uploadErr.Error())
		return "", err
	}

	return uploadRes.Location, nil
}