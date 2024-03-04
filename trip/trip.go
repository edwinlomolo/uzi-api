package trip

import (
	"sync"

	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/location"
	"github.com/edwinlomolo/uzi-api/logger"
	r "github.com/edwinlomolo/uzi-api/repository"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TripService interface {
	FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error)
	GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error)
	AssignCourierToTrip(tripID, courierID uuid.UUID) error
	UnassignTrip(courierID uuid.UUID) error
	CreateTrip(sqlStore.CreateTripParams) (*model.Trip, error)
	CreateTripCost(tripID uuid.UUID, cost int) error
	SetTripStatus(tripID uuid.UUID, status model.TripStatus) error
	MatchCourier(tripID uuid.UUID, pickup model.TripInput)
	CreateTripRecipient(tripID uuid.UUID, input model.TripRecipientInput) error
	GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error)
	GetTrip(tripID uuid.UUID) (*model.Trip, error)
	GetCourierAssignedTrip(courierID uuid.UUID) error
	GetTripCourier(courierID uuid.UUID) (*model.Courier, error)
	ReportTripStatus(tripID uuid.UUID, status model.TripStatus) error
	ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error)
	ParsePickupDropoff(input model.TripInput) (*location.Geocode, error)
}

type tripClient struct {
	r   *r.TripRepository
	mu  sync.Mutex
	log *logrus.Logger
}

func New(store *sqlc.Queries, redis cache.Cache) TripService {
	log := logger.New()
	t := &r.TripRepository{}
	t.Init(store, redis)
	log.Infoln("Trip service...OK")
	return &tripClient{
		t,
		sync.Mutex{},
		log,
	}
}

func (t *tripClient) ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error) {
	return t.r.ComputeTripRoute(input)
}

func (t *tripClient) ParsePickupDropoff(input model.TripInput) (*location.Geocode, error) {
	return t.r.ParsePickupDropoff(input)
}

func (t *tripClient) FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.FindAvailableCourier(pickup)
}

func (t *tripClient) AssignCourierToTrip(tripID, courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.AssignCourierToTrip(tripID, courierID)
}

func (t *tripClient) UnassignTrip(courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.UnassignTrip(courierID)
}

func (t *tripClient) CreateTrip(
	args sqlStore.CreateTripParams,
) (*model.Trip, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.CreateTrip(args)
}

func (t *tripClient) CreateTripCost(tripID uuid.UUID, cost int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.CreateTripCost(tripID, cost)
}

func (t *tripClient) SetTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.SetTripStatus(tripID, status)
}

func (t *tripClient) GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.GetCourierNearPickupPoint(pickup)
}

func (t *tripClient) GetCourierAssignedTrip(courierID uuid.UUID) error {
	return t.r.GetCourierAssignedTrip(courierID)
}

func (t *tripClient) MatchCourier(tripID uuid.UUID, pickup model.TripInput) {
	t.r.MatchCourier(tripID, pickup)
}

func (t *tripClient) CreateTripRecipient(
	tripID uuid.UUID,
	input model.TripRecipientInput,
) error {
	return t.r.CreateTripRecipient(tripID, input)
}

func (t *tripClient) GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error) {
	return t.r.GetTripRecipient(tripID)
}

func (t *tripClient) GetTrip(tripID uuid.UUID) (*model.Trip, error) {
	return t.r.GetTrip(tripID)
}

func (t *tripClient) GetTripCourier(courierID uuid.UUID) (*model.Courier, error) {
	return t.r.GetTripCourier(courierID)
}

func (t *tripClient) ReportTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	return t.r.ReportTripStatus(tripID, status)
}
