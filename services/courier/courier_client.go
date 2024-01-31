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

var courierService Courier

type courierClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
}

func GetCourierService() Courier {
	return courierService
}

func NewCourierService() {
	log := logger.GetLogger()
	courierService = &courierClient{log, store.GetDatabase()}
	log.Infoln("Courier sevice...OK")
}

func (c *courierClient) FindOrCreate(userID uuid.UUID) (*model.Courier, error) {
	courier, err := c.store.GetCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		newCourier, err := c.store.CreateCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
		if err != nil {
			courierErr := model.UziErr{Err: err.Error(), Message: "createcourier", Code: 400}
			c.logger.Errorf("%s: %s", courierErr.Message, courierErr.Err)
			return nil, courierErr
		}

		return &model.Courier{ID: newCourier.ID}, nil
	} else if err != nil {
		courierErr := model.UziErr{Err: err.Error(), Message: "getcourier", Code: 404}
		c.logger.Errorf("%s: %s", courierErr.Message, courierErr.Err)
		return nil, courierErr
	}

	return &model.Courier{ID: courier.ID}, nil
}

func (c *courierClient) IsCourier(userID uuid.UUID) (bool, error) {
	isCourier, err := c.store.IsCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		courierErr := model.UziErr{Err: err.Error(), Message: "checkusercourierstatus", Code: 400}
		c.logger.Errorf("%s: %s", courierErr.Message, courierErr.Err)
		return false, courierErr
	}

	return isCourier.Bool, nil
}

func (c *courierClient) GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error) {
	status, err := c.store.GetCourierStatus(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		return model.CourierStatusOnboarding, nil
	} else if err != nil {
		courierErr := model.UziErr{Err: err.Error(), Message: "getcourierverificationstatus", Code: 500}
		c.logger.Errorf("%s: %s", courierErr.Message, courierErr.Err)
		return model.CourierStatusOffline, courierErr
	}

	return model.CourierStatus(status), nil
}

func (c *courierClient) getCourier(userID uuid.UUID) (*model.Courier, error) {
	var courier model.Courier
	foundCourier, err := c.store.GetCourier(context.Background(), uuid.NullUUID{UUID: userID, Valid: true})
	if err == sql.ErrNoRows {
		noCourierErr := model.UziErr{Err: errors.New("no courier found").Error(), Message: "nocourier", Code: 404}
		c.logger.Errorf(noCourierErr.Error())
		return nil, noCourierErr
	} else if err != nil {
		courierErr := model.UziErr{Err: err.Error(), Message: "getcourier", Code: 500}
		c.logger.Errorf(courierErr.Error())
		return nil, courierErr
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
		uziErr := model.UziErr{Err: setErr.Error(), Message: "setcourierstatus", Code: 500}
		c.logger.Errorf(uziErr.Error())
		return false, uziErr
	}

	return true, nil
}

func (c *courierClient) GetNearbyAvailableProducts(params sqlStore.GetNearbyAvailableCourierProductsParams, tripDistance int) ([]*model.Product, error) {
	var nearbyProducts []*model.Product

	nearbys, nearbyErr := c.store.GetNearbyAvailableCourierProducts(context.Background(), params)
	if nearbyErr == sql.ErrNoRows {
		return make([]*model.Product, 0), nil
	} else if nearbyErr != nil {
		uziErr := model.UziErr{Err: nearbyErr.Error(), Message: "getnearbyavailablecourierproducts", Code: 500}
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	for _, item := range nearbys {
		earnWithFuel := item.Name != "UziX"
		product := &model.Product{
			ID:          item.ID,
			Price:       c.courierProductTripCost(int(item.WeightClass), tripDistance, earnWithFuel),
			Name:        item.Name,
			Description: item.Description,
			IconURL:     item.Icon,
		}

		nearbyProducts = append(nearbyProducts, product)
	}

	return nearbyProducts, nil
}

func (c *courierClient) courierProductTripCost(weightClass, distance int, earnWithFuel bool) int {
	pricer := pricer.GetPricerService()
	p := pricer.CalculateTripCost(weightClass, distance, earnWithFuel)
	return p
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
		uziErr := model.UziErr{Err: err.Error(), Message: "getcouriernearpickuppoint", Code: 500}
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	for _, item := range foundCouriers {
		courier := &model.Courier{
			ID:       item.ID,
			Location: parseCourierLocation(item.Location),
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
