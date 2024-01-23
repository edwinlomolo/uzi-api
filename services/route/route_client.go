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
	_, pickupErr := r.parsePickupDropoff(*input.Pickup)
	if pickupErr != nil {
		return nil, pickupErr
	}

	_, dropoffErr := r.parsePickupDropoff(*input.Dropoff)
	if dropoffErr != nil {
		return nil, dropoffErr
	}

	return nil, nil
}

func (r *routeClient) parsePickupDropoff(input model.TripInput) (model.Geocode, error) {
	geo := &model.Geocode{}

	if input.Location.Lat == 0.0 && input.Location.Lng == 0.0 {
		placedetails, err := location.GetLocationService().GetPlaceDetails(input.PlaceID)
		if err != nil {
			return model.Geocode{}, err
		}

		geo = placedetails
	}

	return *geo, nil
}
