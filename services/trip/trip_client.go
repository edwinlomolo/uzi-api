package trip

import (
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/route"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

type tripClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
	route  route.Route
}

var tripService Trip

func NewTripService(logger *logrus.Logger, store *sqlStore.Queries) {
	tripService = &tripClient{logger, store, route.GetRouteService()}
}

func GetTripService() Trip { return tripService }

func (t *tripClient) ComputeTrip(input model.TripRouteInput) (*model.TripRoute, error) {
	return t.route.ComputeTripRoute(input)
}
