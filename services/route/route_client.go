package route

import (
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/services/location"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

var routeService Route

type routeClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
}

func NewRouteService(logger *logrus.Logger, store *sqlStore.Queries) {
	routeService = &routeClient{logger, store}
}

func GetRouteService() Route { return routeService }

func (r *routeClient) GetTripRoute(input model.TripRouteInput) (*model.TripRoute, error) {
	pickup, pickupErr := r.parsePickupDropoff(*input.Pickup)
	if pickupErr != nil {
		return nil, pickupErr
	}

	dropoff, dropoffErr := r.parsePickupDropoff(*input.Dropoff)
	if dropoffErr != nil {
		return nil, dropoffErr
	}

	return r.computeRoute(*pickup, *dropoff)
}

func (r *routeClient) parsePickupDropoff(input model.TripInput) (*model.Geocode, error) {
	geo := &model.Geocode{}

	// Google place autocomplete select won't have cord in the request
	if input.Location.Lat == 0.0 && input.Location.Lng == 0.0 {
		placedetails, err := location.GetLocationService().GetPlaceDetails(input.PlaceID)
		if err != nil {
			return nil, err
		}

		geo = placedetails
	}

	return geo, nil
}

func (r *routeClient) computeRoute(pickup, dropoff model.Geocode) (*model.TripRoute, error) {
	return nil, nil
}
