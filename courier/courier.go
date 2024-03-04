package courier

import (
	"errors"

	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/gql/model"
	r "github.com/edwinlomolo/uzi-api/repository"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
)

var ErrNoCourierErr = errors.New("no courier found")

type CourierService interface {
	FindOrCreate(userID uuid.UUID) (*model.Courier, error)
	IsCourier(userID uuid.UUID) (bool, error)
	GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error)
	GetCourierByUserID(userID uuid.UUID) (*model.Courier, error)
	GetCourierByID(courierID uuid.UUID) (*model.Courier, error)
	TrackCourierLocation(userID uuid.UUID, input model.GpsInput) error
	UpdateCourierStatus(userID uuid.UUID, status model.CourierStatus) (bool, error)
	GetCourierProduct(productID uuid.UUID) (*model.Product, error)
}

type courierClient struct {
	r *r.CourierRepository
}

func New(store *sqlc.Queries, redis cache.Cache) CourierService {
	cr := &r.CourierRepository{}
	cr.Init(store, redis)
	return &courierClient{cr}
}

func (c *courierClient) FindOrCreate(userID uuid.UUID) (*model.Courier, error) {
	return c.r.FindOrCreate(userID)
}

func (c *courierClient) IsCourier(userID uuid.UUID) (bool, error) {
	return c.r.IsCourier(userID)
}

func (c *courierClient) GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error) {
	return c.r.GetCourierStatus(userID)
}

func (c *courierClient) GetCourierByUserID(userID uuid.UUID) (*model.Courier, error) {
	return c.r.GetCourierByUserID(userID)
}

func (c *courierClient) TrackCourierLocation(userID uuid.UUID, input model.GpsInput) error {
	return c.r.TrackCourierLocation(userID, input)
}

func (c *courierClient) UpdateCourierStatus(userID uuid.UUID, status model.CourierStatus) (bool, error) {
	return c.r.UpdateCourierStatus(userID, status)
}

func (c *courierClient) GetCourierProduct(productID uuid.UUID) (*model.Product, error) {
	return c.r.GetCourierProduct(productID)
}

func (c *courierClient) GetCourierByID(courierID uuid.UUID) (*model.Courier, error) {
	return c.r.GetCourierByID(courierID)
}
