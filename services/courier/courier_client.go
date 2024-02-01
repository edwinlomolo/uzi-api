package courier

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	"github.com/3dw1nM0535/uzi-api/pkg/pricer"
	"github.com/3dw1nM0535/uzi-api/store"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	courierService CourierService
)

type courierClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
}

func GetCourierService() CourierService {
	return courierService
}

func NewCourierService() {
	log := logger.GetLogger()
	courierService = &courierClient{log, store.GetDatabase()}
	log.Infoln("Courier sevice...OK")
}

// TODO something ain't adding up here!
func (c *courierClient) FindOrCreate(userID uuid.UUID) (*model.Courier, error) {
	courier, err := c.store.GetCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		newCourier, err := c.store.CreateCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
		if err != nil {
			c.logger.Errorf(err.Error())
			return nil, fmt.Errorf("%s:%v", "createcourier", err.Error())
		}

		return &model.Courier{ID: newCourier.ID}, nil
	} else if err != nil {
		c.logger.Errorf(err.Error())
		return nil, fmt.Errorf("%s:%v", "getcourier", err.Error())
	}

	return &model.Courier{ID: courier.ID}, nil
}

func (c *courierClient) IsCourier(userID uuid.UUID) (bool, error) {
	isCourier, err := c.store.IsCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		c.logger.Errorf(err.Error())
		return false, fmt.Errorf("%s:%v", "checkusercourierstatus", err.Error())
	}

	return isCourier.Bool, nil
}

func (c *courierClient) GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error) {
	status, err := c.store.GetCourierStatus(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return model.CourierStatusOnboarding, nil
	} else if err != nil {
		c.logger.Errorf(err.Error())
		return model.CourierStatusOffline, fmt.Errorf("%s:%v", "getcourierverificationstatus", err.Error())
	}

	return model.CourierStatus(status), nil
}

func (c *courierClient) getCourier(userID uuid.UUID) (*model.Courier, error) {
	var courier model.Courier
	foundCourier, err := c.store.GetCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		noCourierErr := errors.New("no courier found")
		c.logger.Errorf(noCourierErr.Error())
		return nil, fmt.Errorf("%s:%v", "nocourierfound", noCourierErr.Error())
	} else if err != nil {
		c.logger.Errorf(err.Error())
		return nil, fmt.Errorf("%s:%v", "getcourier", err.Error())
	}

	courier.ID = foundCourier.ID
	courier.UserID = foundCourier.UserID.UUID

	return &courier, nil
}

func (c *courierClient) GetCourier(userID uuid.UUID) (*model.Courier, error) {
	return c.getCourier(userID)
}

func (c *courierClient) TrackCourierLocation(userID uuid.UUID, input model.GpsInput) (bool, error) {
	if _, err := c.getCourier(userID); err != nil {
		return false, err
	}

	args := sqlStore.TrackCourierLocationParams{
		UserID:   uuid.NullUUID{UUID: userID, Valid: true},
		Location: fmt.Sprintf("SRID=4326;POINT(%.8f %.8f)", input.Lng, input.Lat),
	}
	if _, updateErr := c.store.TrackCourierLocation(context.Background(), args); updateErr != nil {
		return false, updateErr
	}

	return true, nil
}

func (c *courierClient) UpdateCourierStatus(userID uuid.UUID, status model.CourierStatus) (bool, error) {
	args := sqlStore.SetCourierStatusParams{
		Status: status.String(),
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
	}
	if _, setErr := c.store.SetCourierStatus(context.Background(), args); setErr != nil {
		c.logger.Errorf(setErr.Error())
		return false, fmt.Errorf("%s:%v", "setcourierstatus", setErr.Error())
	}

	return true, nil
}

func (c *courierClient) GetNearbyAvailableProducts(params sqlStore.GetNearbyAvailableCourierProductsParams, tripDistance int) ([]*model.Product, error) {
	var nearbyProducts []*model.Product

	nearbys, nearbyErr := c.store.GetNearbyAvailableCourierProducts(context.Background(), params)
	if nearbyErr == sql.ErrNoRows {
		return make([]*model.Product, 0), nil
	} else if nearbyErr != nil {
		c.logger.Errorf(nearbyErr.Error())
		return nil, fmt.Errorf("%s:%v", "getnearbyavailablecourierproducts", nearbyErr.Error())
	}

	for _, item := range nearbys {
		earnWithFuel := item.Name != "UziX"
		product := &model.Product{
			ID:          item.ID_2,
			Price:       pricer.GetPricerService().CalculateTripCost(int(item.WeightClass), tripDistance, earnWithFuel),
			Name:        item.Name,
			Description: item.Description,
			IconURL:     item.Icon,
		}

		nearbyProducts = append(nearbyProducts, product)
	}

	return nearbyProducts, nil
}

func (c *courierClient) GetCourierNearPickup(point model.GpsInput) ([]*model.Courier, error) {
	var couriers []*model.Courier

	args := sqlStore.GetCourierNearPickupPointParams{
		Point:  fmt.Sprintf("SRID=4326;POINT(%.8f %.8f)", point.Lng, point.Lat),
		Radius: 2000,
	}
	foundCouriers, err := c.store.GetCourierNearPickupPoint(context.Background(), args)
	if err == sql.ErrNoRows {
		return make([]*model.Courier, 0), nil
	} else if err != nil {
		c.logger.Errorf(err.Error())
		return nil, fmt.Errorf("%s: %v", "getcouriernearpickuppooint", err.Error())
	}

	for _, item := range foundCouriers {
		courier := &model.Courier{
			ID:        item.ID,
			ProductID: item.ProductID.UUID,
			Location:  parseCourierLocation(item.Location),
		}

		couriers = append(couriers, courier)
	}

	return couriers, nil
}

func parseCourierLocation(point interface{}) *model.Gps {
	var location *model.Point

	if point != nil {
		json.Unmarshal([]byte((point).(string)), &location)

		lat := &location.Coordinates[1]
		lng := &location.Coordinates[0]
		return &model.Gps{
			Lat: *lat,
			Lng: *lng,
		}
	} else {
		return nil
	}
}

func (c *courierClient) GetCourierProduct(id uuid.UUID) (*model.Product, error) {
	product, err := c.store.GetCourierProductByID(context.Background(), id)
	if err != nil {
		c.logger.Errorf(err.Error())
		return nil, fmt.Errorf("%s:%v", "get courier product", err.Error())
	}

	return &model.Product{
		ID:      product.ID,
		IconURL: product.Icon,
		Name:    product.Name,
	}, nil
}
