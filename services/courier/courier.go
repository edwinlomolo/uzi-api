package courier

import (
	"github.com/3dw1nM0535/uzi-api/model"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/google/uuid"
)

type CourierService interface {
	FindOrCreate(userID uuid.UUID) (*model.Courier, error)
	IsCourier(userID uuid.UUID) (bool, error)
	GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error)
	GetCourier(userID uuid.UUID) (*model.Courier, error)
	TrackCourierLocation(userID uuid.UUID, input model.GpsInput) (bool, error)
	UpdateCourierStatus(userID uuid.UUID, status model.CourierStatus) (bool, error)
	GetNearbyAvailableProducts(params sqlStore.GetNearbyAvailableCourierProductsParams, tripDistance int) ([]*model.Product, error)
	GetCourierNearPickup(point model.GpsInput) ([]*model.Courier, error)
	GetCourierProduct(product_id uuid.UUID) (*model.Product, error)
}
