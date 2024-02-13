package courier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrNoCourierErr = errors.New("no courier found")
	Courier         CourierService
)

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
	logger *logrus.Logger
	store  *sqlStore.Queries
}

func NewCourierService() {
	Courier = &courierClient{logger.Logger, store.DB}
	logger.Logger.Infoln("Courier sevice...OK")
}

func (c *courierClient) FindOrCreate(userID uuid.UUID) (*model.Courier, error) {
	courier, err := c.getCourier(userID)
	if err == nil && courier == nil {
		newCourier, newErr := c.store.CreateCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
		if newErr != nil {
			uziErr := fmt.Errorf("%s:%v", "create courier", newErr)
			c.logger.Errorf(uziErr.Error())
			return nil, uziErr
		}

		return &model.Courier{ID: newCourier.ID}, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier", err)
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Courier{ID: courier.ID}, nil
}

func (c *courierClient) IsCourier(userID uuid.UUID) (bool, error) {
	isCourier, err := c.store.IsCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "check user courier status", err.Error())
		c.logger.Errorf(uziErr.Error())
		return false, uziErr
	}

	return isCourier.Bool, nil
}

func (c *courierClient) GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error) {
	status, err := c.store.GetCourierStatus(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return model.CourierStatusOnboarding, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier verification status", err.Error())
		c.logger.Errorf(uziErr.Error())
		return model.CourierStatusOffline, uziErr
	}

	return model.CourierStatus(status), nil
}

func (c *courierClient) getCourier(userID uuid.UUID) (*model.Courier, error) {
	var courier model.Courier
	foundCourier, err := c.store.GetCourierByUserID(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier", err)
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	courier.ID = foundCourier.ID
	courier.UserID = foundCourier.UserID.UUID
	courier.Avatar = c.getAvatar(foundCourier.ID)

	return &courier, nil
}

func (c *courierClient) getAvatar(courierID uuid.UUID) *model.Uploads {
	ID := uuid.NullUUID{UUID: courierID, Valid: true}
	avatar, err := c.store.GetCourierAvatar(context.Background(), ID)
	if err != nil && err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier avatar", err)
		c.logger.Errorf(uziErr.Error())
		return nil
	}

	return &model.Uploads{ID: avatar.ID, URI: avatar.Uri}
}

func (c *courierClient) GetCourierByUserID(userID uuid.UUID) (*model.Courier, error) {
	return c.getCourier(userID)
}

func (c *courierClient) TrackCourierLocation(userID uuid.UUID, input model.GpsInput) error {
	if _, err := c.getCourier(userID); err != nil {
		return err
	}

	args := sqlStore.TrackCourierLocationParams{
		UserID:   uuid.NullUUID{UUID: userID, Valid: true},
		Location: fmt.Sprintf("SRID=4326;POINT(%.8f %.8f)", input.Lng, input.Lat),
	}
	if _, updateErr := c.store.TrackCourierLocation(context.Background(), args); updateErr != nil {
		uziErr := fmt.Errorf("%s:%v", "track courier location", updateErr)
		c.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (c *courierClient) UpdateCourierStatus(userID uuid.UUID, status model.CourierStatus) (bool, error) {
	args := sqlStore.SetCourierStatusParams{
		Status: status.String(),
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
	}
	if _, setErr := c.store.SetCourierStatus(context.Background(), args); setErr != nil {
		uziErr := fmt.Errorf("%s:%v", "set courier status", setErr.Error())
		c.logger.Errorf(uziErr.Error())
		return false, uziErr
	}

	return true, nil
}

func (c *courierClient) GetCourierProduct(productID uuid.UUID) (*model.Product, error) {
	product, err := c.store.GetCourierProductByID(context.Background(), productID)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier product", err.Error())
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Product{
		ID:      product.ID,
		IconURL: product.Icon,
		Name:    product.Name,
	}, nil
}

func (c *courierClient) GetCourierByID(courierID uuid.UUID) (*model.Courier, error) {
	courier, err := c.store.GetCourierByID(context.Background(), courierID)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier by id", err)
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Courier{ID: courier.ID, TripID: &courier.TripID.UUID}, nil
}
