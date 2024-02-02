package upload

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/3dw1nM0535/uzi-api/internal/aws"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/store"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var uploadService Upload

func GetUploadService() Upload { return uploadService }

type uploadClient struct {
	s3     aws.Aws
	logger *logrus.Logger
	store  *sqlStore.Queries
}

func NewUploadService() {
	log := logger.GetLogger()
	uploadService = &uploadClient{aws.GetAwsService(), log, store.GetDatabase()}
	log.Infoln("Upload service...OK")
}

func (u *uploadClient) CreateCourierUpload(reason, uri string, id uuid.UUID) error {
	return u.createCourierUpload(reason, uri, id)
}

func (u *uploadClient) CreateUserUpload(reason, uri string, id uuid.UUID) error {
	return u.createUserUpload(reason, uri, id)
}

func (u *uploadClient) createCourierUpload(reason, uri string, id uuid.UUID) error {
	courierArgs := sqlStore.GetCourierUploadParams{
		Type:      reason,
		CourierID: uuid.NullUUID{UUID: id, Valid: true},
	}

	courierUpload, getErr := u.store.GetCourierUpload(context.Background(), courierArgs)
	if getErr == sql.ErrNoRows {
		createArgs := sqlStore.CreateCourierUploadParams{
			Type:         reason,
			Uri:          uri,
			CourierID:    uuid.NullUUID{UUID: id, Valid: true},
			Verification: model.UploadVerificationStatusVerifying.String(),
		}

		_, createErr := u.store.CreateCourierUpload(context.Background(), createArgs)
		if createErr != nil {
			uziErr := fmt.Errorf("%s:%v", "create courier upload", createErr)
			u.logger.Errorf(uziErr.Error())
			return uziErr
		}

		return nil
	} else if getErr != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier upload", getErr)
		u.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return u.updateUploadUri(courierUpload.Uri, courierUpload.ID)
}

func (u *uploadClient) updateUploadUri(uri string, ID uuid.UUID) error {
	updateParams := sqlStore.UpdateUploadParams{
		ID:           ID,
		Uri:          sql.NullString{String: uri, Valid: true},
		Verification: sql.NullString{String: model.UploadVerificationStatusVerifying.String(), Valid: true},
	}

	if _, updateErr := u.store.UpdateUpload(context.Background(), updateParams); updateErr != nil {
		uziErr := fmt.Errorf("%s:%v", "update courier upload", updateErr)
		u.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (u *uploadClient) updateUploadVerificationStatus(id uuid.UUID, status model.UploadVerificationStatus) error {
	args := sqlStore.UpdateUploadParams{
		ID:           id,
		Verification: sql.NullString{String: status.String(), Valid: true},
	}
	if _, updateErr := u.store.UpdateUpload(context.Background(), args); updateErr != nil {
		err := fmt.Errorf("%s:%v", "set upload verification", updateErr)
		u.logger.Errorf(err.Error())
		return err
	}

	return nil
}

func (u *uploadClient) createUserUpload(reason, uri string, ID uuid.UUID) error {
	createParams := sqlStore.CreateUserUploadParams{
		Type:   reason,
		Uri:    uri,
		UserID: uuid.NullUUID{UUID: ID, Valid: true},
	}

	foundUpload, foundErr := u.store.CreateUserUpload(context.Background(), createParams)
	if foundErr == sql.ErrNoRows {
		_, createErr := u.store.CreateUserUpload(context.Background(), createParams)
		if createErr != nil {
			uziErr := fmt.Errorf("%s:%v", "create user upload", createErr)
			u.logger.Errorf(uziErr.Error())
			return uziErr
		}
	} else if foundErr != nil {
		uziErr := fmt.Errorf("%s:%v", "get user upload", foundErr)
		u.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return u.updateUploadUri(foundUpload.Uri, foundUpload.ID)
}

func (u *uploadClient) GetCourierUploads(courierID uuid.UUID) ([]*model.Uploads, error) {
	var uploads []*model.Uploads

	args := uuid.NullUUID{UUID: courierID, Valid: true}
	uplds, uploadsErr := u.store.GetCourierUploads(context.Background(), args)
	if uploadsErr != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier uploads", uploadsErr)
		u.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	for _, i := range uplds {
		upload := &model.Uploads{
			ID:           i.ID,
			URI:          i.Uri,
			Type:         i.Type,
			Verification: model.UploadVerificationStatus(i.Verification),
		}

		uploads = append(uploads, upload)
	}

	return uploads, nil
}
