package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/repository"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrNoCourierErr        = errors.New("courier service: no courier found")
	ErrCourierTripNotFound = errors.New("courier service: courier trip not found")
	cService               CourierService
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
	r     *repository.CourierRepository
	log   *logrus.Logger
	cache internal.Cache
}

func NewCourierService() {
	cr := &repository.CourierRepository{}
	cr.Init()
	cService = &courierClient{
		cr,
		internal.GetLogger(),
		internal.GetCache(),
	}
}

func GetCourierService() CourierService {
	return cService
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
	if err := c.r.TrackCourierLocation(userID, input); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		courier, err := c.r.GetCourierByUserID(userID)
		if err != nil {
			return
		}

		t, err := c.r.GetCourierTrip(courier.ID)
		if err != nil && errors.Is(err, ErrCourierTripNotFound) {
			c.log.WithError(ErrCourierTripNotFound).Errorf("is courier tripping")
			return
		} else if err != nil {
			return
		}

		if t != nil && (t.Status == model.TripStatusCourierEnRoute || t.Status == model.TripStatusCourierArriving) {
			tripUpdate := model.TripUpdate{
				ID:     t.ID,
				Status: model.TripStatus(t.Status),
				Location: &model.Gps{
					Lat: input.Lat,
					Lng: input.Lng,
				},
			}
			u, marshalErr := json.Marshal(tripUpdate)
			if marshalErr != nil {
				c.log.WithError(marshalErr).Errorf("courier service: marshal courier arriving/enroute trip update")
				return
			}
			tripUpdateErr := c.cache.GetRedis().Publish(context.Background(), internal.TRIP_UPDATES_CHANNEL, u).Err()
			if tripUpdateErr != nil {
				c.log.WithFields(logrus.Fields{
					"status":  t.Status,
					"trip_id": t.ID,
				}).WithError(tripUpdateErr).Errorf("courier service: publish courier arriving/enroute trip update")
				return
			}
		}
	}()
	<-done

	return nil
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
