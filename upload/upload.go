package upload

import (
	"github.com/edwinlomolo/uzi-api/gql/model"
	r "github.com/edwinlomolo/uzi-api/repository"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UploadService interface {
	CreateCourierUpload(reason, uri string, courierID uuid.UUID) error
	CreateUserUpload(reason, uri string, userID uuid.UUID) error
	GetCourierUploads(courierID uuid.UUID) ([]*model.Uploads, error)
}

type uploadClient struct {
	r r.UploadRepository
}

func New(store *sqlc.Queries, redis *redis.Client) UploadService {
	ur := r.UploadRepository{}
	ur.Init(store)
	return &uploadClient{ur}
}

func (u *uploadClient) CreateCourierUpload(reason, uri string, id uuid.UUID) error {
	return u.r.CreateCourierUpload(reason, uri, id)
}

func (u *uploadClient) CreateUserUpload(reason, uri string, id uuid.UUID) error {
	return u.r.CreateUserUpload(reason, uri, id)
}

func (u *uploadClient) GetCourierUploads(courierID uuid.UUID) ([]*model.Uploads, error) {
	return u.r.GetCourierUploads(courierID)
}
