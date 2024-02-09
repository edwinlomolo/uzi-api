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
	GetCourierByID(userID uuid.UUID) (*model.Courier, error)
	TrackCourierLocation(userID uuid.UUID, input model.GpsInput) error
	UpdateCourierStatus(userID uuid.UUID, status model.CourierStatus) (bool, error)
	GetCourierProduct(product_id uuid.UUID) (*model.Product, error)
}

type courierClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
}

func NewCourierService() {
	Courier = &courierClient{logger.Logger, store.DB}
	logger.Logger.Infoln("Courier sevice...OK")
}

// TODO something ain't adding up here!
func (c *courierClient) FindOrCreate(userID uuid.UUID) (*model.Courier, error) {
	courier, err := c.store.GetCourierByUserID(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		newCourier, err := c.store.CreateCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
		if err != nil {
			uziErr := fmt.Errorf("%s:%v", "createcourier", err.Error())
			c.logger.Errorf(uziErr.Error())
			return nil, uziErr
		}

		return &model.Courier{ID: newCourier.ID}, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "getcourier", err.Error())
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
		uziErr := fmt.Errorf("%s:%v", "checkusercourierstatus", err.Error())
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
		uziErr := fmt.Errorf("%s:%v", "getcourierverificationstatus", err.Error())
		c.logger.Errorf(uziErr.Error())
		return model.CourierStatusOffline, uziErr
	}

	return model.CourierStatus(status), nil
}

func (c *courierClient) getCourier(userID uuid.UUID) (*model.Courier, error) {
	var courier model.Courier
	foundCourier, err := c.store.GetCourierByUserID(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		uziErr := fmt.Errorf("%s:%v", "nocourierfound", ErrNoCourierErr.Error())
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "getcourier", err.Error())
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	courier.ID = foundCourier.ID
	courier.UserID = foundCourier.UserID.UUID

	return &courier, nil
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
		uziErr := fmt.Errorf("%s:%v", "trackcourierlocation", updateErr)
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
		uziErr := fmt.Errorf("%s:%v", "setcourierstatus", setErr.Error())
		c.logger.Errorf(uziErr.Error())
		return false, uziErr
	}

	return true, nil
}

func (c *courierClient) GetCourierProduct(id uuid.UUID) (*model.Product, error) {
	product, err := c.store.GetCourierProductByID(context.Background(), id)
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
