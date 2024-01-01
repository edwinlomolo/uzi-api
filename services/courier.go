package services

import (
	"context"
	"database/sql"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var courierService Courier

type Courier interface {
	FindOrCreate(userID uuid.UUID) (*model.Courier, error)
	IsCourier(userID uuid.UUID) (bool, error)
	GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error)
}

type courierClient struct {
	logger *logrus.Logger
	store  *store.Queries
}

func GetCourierService() Courier {
	return courierService
}

func NewCourierService(logger *logrus.Logger, store *store.Queries) Courier {
	courierService = &courierClient{logger, store}
	return courierService
}

func (c *courierClient) FindOrCreate(userID uuid.UUID) (*model.Courier, error) {
	courier, err := c.store.GetCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		newCourier, err := c.store.CreateCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
		if err != nil {
			c.logger.Errorf("%s-%v", "CreateCourierErr", err.Error())
			return nil, err
		}

		return &model.Courier{ID: newCourier.ID}, nil
	} else if err != nil {
		c.logger.Errorf("%s-%v", "CreateCourierErr", err.Error())
		return nil, err
	}

	return &model.Courier{ID: courier.ID}, nil
}

func (c *courierClient) IsCourier(userID uuid.UUID) (bool, error) {
	isCourier, err := c.store.IsCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		c.logger.Errorf("%s-%v", "IsCourierErr", err.Error())
		return false, err
	}

	return isCourier.Bool, nil
}

func (c *courierClient) GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error) {
	status, err := c.store.GetCourierStatus(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return model.CourierStatusOffline, nil
	} else if err != nil {
		c.logger.Errorf("%s-%v", "GetCourierStatusErr", err.Error())
		return model.CourierStatusOffline, err
	}

	return model.CourierStatus(status), nil
}
