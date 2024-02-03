package trip

import (
	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/internal/route"
	"github.com/3dw1nM0535/uzi-api/store"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

type TripService interface {
	ComputeTrip(input model.TripRouteInput) (*model.TripRoute, error)
}

type tripClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
	route  route.Route
}

var Trip TripService

func NewTripService() {
	Trip = &tripClient{logger.Logger, store.DB, route.Routing}
	logger.Logger.Infoln("Trip service...OK")
}

func (t *tripClient) ComputeTrip(input model.TripRouteInput) (*model.TripRoute, error) {
	return t.route.ComputeTripRoute(input)
}
