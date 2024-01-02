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
	FindOrCreate(userID uuid.UUID) (*model.Courier, *model.UziErr)
	IsCourier(userID uuid.UUID) (bool, *model.UziErr)
	GetCourierStatus(userID uuid.UUID) (model.CourierStatus, *model.UziErr)
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

func (c *courierClient) FindOrCreate(userID uuid.UUID) (*model.Courier, *model.UziErr) {
	courier, err := c.store.GetCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		newCourier, err := c.store.CreateCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
		if err != nil {
			courierErr := &model.UziErr{Error: err, Message: "create courier error", Code: 400}
			c.logger.Errorf(courierErr.ErrorString())
			return nil, courierErr
		}

		return &model.Courier{ID: newCourier.ID}, nil
	} else if err != nil {
		courierErr := &model.UziErr{Error: err, Message: "get courier error", Code: 404}
		c.logger.Errorf(courierErr.ErrorString())
		return nil, courierErr
	}

	return &model.Courier{ID: courier.ID}, nil
}

func (c *courierClient) IsCourier(userID uuid.UUID) (bool, *model.UziErr) {
	isCourier, err := c.store.IsCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		courierErr := &model.UziErr{Error: err, Message: "check user courier status err", Code: 400}
		c.logger.Errorf(courierErr.ErrorString())
		return false, courierErr
	}

	return isCourier.Bool, nil
}

func (c *courierClient) GetCourierStatus(userID uuid.UUID) (model.CourierStatus, *model.UziErr) {
	status, err := c.store.GetCourierStatus(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return model.CourierStatusOffline, nil
	} else if err != nil {
		courierErr := &model.UziErr{Error: err, Message: "get courier verification status error", Code: 500}
		c.logger.Errorf(courierErr.ErrorString())
		return model.CourierStatusOffline, courierErr
	}

	return model.CourierStatus(status), nil
}
