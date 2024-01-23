package trip

import (
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

type tripClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
}

var tripService Trip

func NewTripService(logger *logrus.Logger, store *sqlStore.Queries) {
	tripService = &tripClient{logger, store}
}

func GetTripService() Trip { return tripService }

func (t *tripClient) CreateTrip() {}

func (t *tripClient) AssingRouteToTrip() {}
