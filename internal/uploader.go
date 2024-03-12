package internal

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"

	"cloud.google.com/go/storage"
	"github.com/edwinlomolo/uzi-api/config"
	"google.golang.org/api/option"
)

var (
	gcs GCS
)

type GCS interface {
	UploadCourierDocument(multipart.File, *multipart.FileHeader) (string, error)
}

type gcsClient struct {
	cStorage      *storage.Client
	courierBucket string
}

func NewUploader() {
	credentials, err := base64.StdEncoding.DecodeString(config.Config.Google.GoogleApplicationDevelopmentCredentials)
	if err != nil {
		log.WithError(err).Fatalln("reading google ADC")
	}
	storage, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(credentials))
	if err != nil {
		log.WithError(err).Fatalln("new google cloud storage client")
	}

	gcs = &gcsClient{storage, config.Config.Google.GoogleCloudStorageCourierDocumentsBucket}
}

func GetGCS() GCS {
	return gcs
}

func (cs *gcsClient) UploadCourierDocument(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	bucket := config.Config.Google.GoogleCloudStorageCourierDocumentsBucket
	sw := cs.cStorage.Bucket(bucket).Object(fileHeader.Filename).NewWriter(context.Background())

	if _, err := io.Copy(sw, file); err != nil {
		log.WithField("file_size", fileHeader.Size).WithError(err).Errorf("courier document upload")
		return "", err
	}

	if err := sw.Close(); err != nil {
		log.WithError(err).Errorf("closing google cloud storage object writer")
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", config.Config.Google.GoogleCloudObjectUri, bucket, fileHeader.Filename), nil
}
