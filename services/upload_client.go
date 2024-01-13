package services

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/aws"
	"github.com/3dw1nM0535/uzi-api/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var uploadService Upload

func GetUploadService() Upload { return uploadService }

type uploadClient struct {
	s3     aws.Aws
	logger *logrus.Logger
	store  *store.Queries
}

func NewUploadService(s3 aws.Aws, logger *logrus.Logger, store *store.Queries) Upload {
	uploadService = &uploadClient{s3, logger, store}

	logger.Infoln("Upload service...OK")
	return uploadService
}

func (u *uploadClient) CreateCourierUpload(reason, uri string, id uuid.UUID) error {
	return u.createCourierUpload(reason, uri, id)
}

func (u *uploadClient) CreateUserUpload(reason, uri string, id uuid.UUID) error {
	return u.createUserUpload(reason, uri, id)
}

func (u *uploadClient) createCourierUpload(reason, uri string, id uuid.UUID) error {
	courierArgs := store.GetCourierUploadParams{
		Type:      reason,
		CourierID: uuid.NullUUID{UUID: id, Valid: true},
	}

	courierUpload, getErr := u.store.GetCourierUpload(context.Background(), courierArgs)
	if getErr == sql.ErrNoRows {
		createArgs := store.CreateCourierUploadParams{
			Type:      reason,
			Uri:       uri,
			CourierID: uuid.NullUUID{UUID: id, Valid: true},
		}

		_, createErr := u.store.CreateCourierUpload(context.Background(), createArgs)
		if createErr != nil {
			uziErr := model.UziErr{Err: createErr.Error(), Message: "createcourierupload", Code: 500}
			u.logger.Errorf(uziErr.Error())
			return createErr
		}

		return nil
	} else if getErr != nil {
		uziErr := model.UziErr{Err: getErr.Error(), Message: "getcourierupload", Code: 500}
		u.logger.Errorf(uziErr.Error())
		return getErr
	}

	return u.updateUpload(uri, courierUpload.ID)
}

func (u *uploadClient) updateUpload(uri string, ID uuid.UUID) error {
	updateParams := store.UpdateUploadParams{
		ID:  ID,
		Uri: uri,
	}

	_, updateErr := u.store.UpdateUpload(context.Background(), updateParams)
	if updateErr != nil {
		uziErr := model.UziErr{Err: updateErr.Error(), Message: "updateupload", Code: 500}
		u.logger.Errorf(uziErr.Error())
		return updateErr
	}

	return nil
}

func (u *uploadClient) createUserUpload(reason, uri string, ID uuid.UUID) error {
	createParams := store.CreateUserUploadParams{
		Type:   reason,
		Uri:    uri,
		UserID: uuid.NullUUID{UUID: ID, Valid: true},
	}

	foundUpload, foundErr := u.store.CreateUserUpload(context.Background(), createParams)
	if foundErr == sql.ErrNoRows {
		_, createErr := u.store.CreateUserUpload(context.Background(), createParams)
		if createErr != nil {
			uziErr := model.UziErr{Err: createErr.Error(), Message: "createuserupload", Code: 500}
			u.logger.Errorf(uziErr.Error())
			return createErr
		}
	} else if foundErr != nil {
		uziErr := model.UziErr{Err: foundErr.Error(), Message: "getuserupload", Code: 500}
		u.logger.Errorf(uziErr.Error())
		return foundErr
	}

	return u.updateUpload(uri, foundUpload.ID)
}

func (u *uploadClient) GetCourierUploads(courierID uuid.UUID) ([]*model.Uploads, error) {
	var uploads []*model.Uploads

	args := uuid.NullUUID{UUID: courierID, Valid: true}
	uplds, uploadsErr := u.store.GetCourierUploads(context.Background(), args)
	if uploadsErr != nil {
		uziErr := model.UziErr{Err: uploadsErr.Error(), Message: "getcourieruploads", Code: http.StatusInternalServerError}
		u.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	for _, i := range uplds {
		upload := &model.Uploads{
			ID:       i.ID,
			URI:      &i.Uri,
			Type:     &i.Type,
			Verified: i.Verified,
		}

		uploads = append(uploads, upload)
	}

	return uploads, nil
}
