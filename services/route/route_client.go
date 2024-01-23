package route

import (
	"github.com/3dw1nM0535/uzi-api/model"
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
	return nil, nil
}
