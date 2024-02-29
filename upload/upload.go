package upload

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/edwinlomolo/uzi-api/aws"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	Upload UploadService
)

type UploadService interface {
	CreateCourierUpload(reason, uri string, courierID uuid.UUID) error
	CreateUserUpload(reason, uri string, userID uuid.UUID) error
	GetCourierUploads(courierID uuid.UUID) ([]*model.Uploads, error)
}

type uploadClient struct {
	s3    aws.Aws
	log   *logrus.Logger
	store *sqlStore.Queries
}

func NewUploadService() {
	Upload = &uploadClient{
		aws.S3,
		logger.Logger,
		store.DB,
	}
	logger.Logger.Infoln("Upload service...OK")
}

func (u *uploadClient) CreateCourierUpload(reason, uri string, id uuid.UUID) error {
	return u.createCourierUpload(reason, uri, id)
}

func (u *uploadClient) CreateUserUpload(reason, uri string, id uuid.UUID) error {
	return u.createUserUpload(reason, uri, id)
}

func (u *uploadClient) createCourierUpload(reason, uri string, id uuid.UUID) error {
	courierArgs := sqlStore.GetCourierUploadParams{
		Type: reason,
		CourierID: uuid.NullUUID{
			UUID:  id,
			Valid: true,
		},
	}

	courierUpload, getErr := u.store.GetCourierUpload(context.Background(), courierArgs)
	if getErr == sql.ErrNoRows {
		createArgs := sqlStore.CreateCourierUploadParams{
			Type: reason,
			Uri:  uri,
			CourierID: uuid.NullUUID{
				UUID:  id,
				Valid: true,
			},
			Verification: model.UploadVerificationStatusVerifying.String(),
		}

		_, createErr := u.store.CreateCourierUpload(context.Background(), createArgs)
		if createErr != nil {
			u.log.WithFields(logrus.Fields{
				"error":      createErr,
				"type":       createArgs.Type,
				"courier_id": createArgs.CourierID.UUID,
			}).Errorf("create courier upload")
			return createErr
		}

		return nil
	} else if getErr != nil {
		u.log.WithFields(logrus.Fields{
			"error":      getErr,
			"courier_id": courierArgs.CourierID.UUID,
			"type":       courierArgs.Type,
		}).Errorf("get courier upload")
		return getErr
	}

	return u.updateUploadUri(courierUpload.Uri, courierUpload.ID)
}

func (u *uploadClient) updateUploadUri(uri string, ID uuid.UUID) error {
	updateParams := sqlStore.UpdateUploadParams{
		ID: ID,
		Uri: sql.NullString{
			String: uri,
			Valid:  true,
		},
		Verification: sql.NullString{
			String: model.UploadVerificationStatusVerifying.String(),
			Valid:  true,
		},
	}

	if _, updateErr := u.store.UpdateUpload(
		context.Background(),
		updateParams); updateErr != nil {
		u.log.WithFields(logrus.Fields{
			"error":     updateErr,
			"upload_id": ID,
		}).Errorf("update upload")
		return updateErr
	}

	return nil
}

func (u *uploadClient) updateUploadVerificationStatus(
	id uuid.UUID,
	status model.UploadVerificationStatus,
) error {
	args := sqlStore.UpdateUploadParams{
		ID: id,
		Verification: sql.NullString{
			String: status.String(),
			Valid:  true,
		},
	}
	if _, updateErr := u.store.UpdateUpload(context.Background(), args); updateErr != nil {
		u.log.WithFields(logrus.Fields{
			"error":     updateErr,
			"upload_id": id,
			"status":    status.String(),
		}).Errorf("update upload verification status")
		return updateErr
	}

	return nil
}

func (u *uploadClient) createUserUpload(reason, uri string, ID uuid.UUID) error {
	getParams := sqlStore.GetUserUploadParams{
		Type: reason,
		UserID: uuid.NullUUID{
			UUID:  ID,
			Valid: true,
		},
	}

	foundUpload, foundErr := u.store.GetUserUpload(context.Background(), getParams)
	if foundErr == sql.ErrNoRows {
		createParams := sqlStore.CreateUserUploadParams{
			Type: reason,
			Uri:  uri,
			UserID: uuid.NullUUID{
				UUID:  ID,
				Valid: true,
			},
		}
		_, createErr := u.store.CreateUserUpload(context.Background(), createParams)
		if createErr != nil {
			u.log.WithFields(logrus.Fields{
				"error":   createErr,
				"type":    createParams.Type,
				"user_id": createParams.UserID.UUID,
			}).Errorf("create user upload")
			return createErr
		}
	} else if foundErr != nil {
		uziErr := fmt.Errorf("%s:%v", "user upload", foundErr)
		u.log.WithFields(logrus.Fields{
			"error":   foundErr,
			"type":    getParams.Type,
			"user_id": getParams.UserID.UUID,
		}).Errorf("found user upload")
		return uziErr
	}

	return u.updateUploadUri(foundUpload.Uri, foundUpload.ID)
}

func (u *uploadClient) GetCourierUploads(
	courierID uuid.UUID,
) ([]*model.Uploads, error) {
	var uploads []*model.Uploads

	args := uuid.NullUUID{UUID: courierID, Valid: true}
	uplds, uploadsErr := u.store.GetCourierUploads(context.Background(), args)
	if uploadsErr != nil {
		u.log.WithFields(logrus.Fields{
			"error":      uploadsErr,
			"courier_id": courierID,
		}).Errorf("get courier uploads")
		return nil, uploadsErr
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
