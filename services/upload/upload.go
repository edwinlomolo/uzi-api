package upload

import (
	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/google/uuid"
)

type Upload interface {
	CreateCourierUpload(reason, uri string, courierID uuid.UUID) error
	CreateUserUpload(reason, uri string, userID uuid.UUID) error
	GetCourierUploads(courierID uuid.UUID) ([]*model.Uploads, error)
}
