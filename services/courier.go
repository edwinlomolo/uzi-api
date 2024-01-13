package services

import (
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/google/uuid"
)

type Courier interface {
	FindOrCreate(userID uuid.UUID) (*model.Courier, error)
	IsCourier(userID uuid.UUID) (bool, error)
	GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error)
	GetCourier(userID uuid.UUID) (*model.Courier, error)
}
