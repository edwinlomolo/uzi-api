package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UploadRepository struct {
	store *sqlc.Queries
	log   *logrus.Logger
}

func (u *UploadRepository) Init(q *sqlc.Queries) {
	u.store = q
	u.log = internal.GetLogger()
}

func (u *UploadRepository) CreateCourierUpload(reason, uri string, id uuid.UUID) error {
	return u.createCourierUpload(reason, uri, id)
}

func (u *UploadRepository) CreateUserUpload(reason, uri string, id uuid.UUID) error {
	return u.createUserUpload(reason, uri, id)
}

func (u *UploadRepository) createCourierUpload(reason, uri string, id uuid.UUID) error {
	courierArgs := sqlc.GetCourierUploadParams{
		Type: reason,
		CourierID: uuid.NullUUID{
			UUID:  id,
			Valid: true,
		},
	}

	courierUpload, getErr := u.store.GetCourierUpload(context.Background(), courierArgs)
	if getErr == sql.ErrNoRows {
		createArgs := sqlc.CreateCourierUploadParams{
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
				"type":       createArgs.Type,
				"courier_id": createArgs.CourierID.UUID,
			}).WithError(createErr).Errorf("create courier upload")
			return createErr
		}

		return nil
	} else if getErr != nil {
		u.log.WithFields(logrus.Fields{
			"courier_id": courierArgs.CourierID.UUID,
			"type":       courierArgs.Type,
		}).WithError(getErr).Errorf("get courier upload")
		return getErr
	}

	return u.updateUploadUri(courierUpload.Uri, courierUpload.ID)
}

func (u *UploadRepository) updateUploadUri(uri string, ID uuid.UUID) error {
	updateParams := sqlc.UpdateUploadParams{
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
			"upload_id": ID,
		}).WithError(updateErr).Errorf("update upload")
		return updateErr
	}

	return nil
}

func (u *UploadRepository) updateUploadVerificationStatus(
	id uuid.UUID,
	status model.UploadVerificationStatus,
) error {
	args := sqlc.UpdateUploadParams{
		ID: id,
		Verification: sql.NullString{
			String: status.String(),
			Valid:  true,
		},
	}
	if _, updateErr := u.store.UpdateUpload(context.Background(), args); updateErr != nil {
		u.log.WithFields(logrus.Fields{
			"upload_id": id,
			"status":    status.String(),
		}).WithError(updateErr).Errorf("update upload verification status")
		return updateErr
	}

	return nil
}

func (u *UploadRepository) createUserUpload(reason, uri string, ID uuid.UUID) error {
	getParams := sqlc.GetUserUploadParams{
		Type: reason,
		UserID: uuid.NullUUID{
			UUID:  ID,
			Valid: true,
		},
	}

	foundUpload, foundErr := u.store.GetUserUpload(context.Background(), getParams)
	if foundErr == sql.ErrNoRows {
		createParams := sqlc.CreateUserUploadParams{
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
				"type":    createParams.Type,
				"user_id": createParams.UserID.UUID,
			}).WithError(createErr).Errorf("create user upload")
			return createErr
		}
	} else if foundErr != nil {
		uziErr := fmt.Errorf("%s:%v", "user upload", foundErr)
		u.log.WithFields(logrus.Fields{
			"type":    getParams.Type,
			"user_id": getParams.UserID.UUID,
		}).WithError(foundErr).Errorf("found user upload")
		return uziErr
	}

	return u.updateUploadUri(foundUpload.Uri, foundUpload.ID)
}

func (u *UploadRepository) GetCourierUploads(
	courierID uuid.UUID,
) ([]*model.Uploads, error) {
	var uploads []*model.Uploads

	args := uuid.NullUUID{UUID: courierID, Valid: true}
	uplds, uploadsErr := u.store.GetCourierUploads(context.Background(), args)
	if uploadsErr != nil {
		u.log.WithFields(logrus.Fields{
			"courier_id": courierID,
		}).WithError(uploadsErr).Errorf("get courier uploads")
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
