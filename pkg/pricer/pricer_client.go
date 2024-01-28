package pricer

import (
	"github.com/3dw1nM0535/uzi-api/config"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

var pricerService Pricer

type pricerClient struct {
	store  *sqlStore.Queries
	logger *logrus.Logger
	config config.Pricer
}

func NewPricer(store *sqlStore.Queries, logger *logrus.Logger, cfg config.Pricer) Pricer {
	pricerService = &pricerClient{store, logger, cfg}

	return pricerService
}
