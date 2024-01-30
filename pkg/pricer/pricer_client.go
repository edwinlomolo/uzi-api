package pricer

import (
	"math"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	"github.com/3dw1nM0535/uzi-api/store"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

var pricerService Pricer

type pricerClient struct {
	store  *sqlStore.Queries
	logger *logrus.Logger
	config config.Pricer
}

func GetPricerService() Pricer { return pricerService }

func NewPricer() {
	pricerService = &pricerClient{store.GetDatabase(), logger.GetLogger(), config.GetConfig().Pricer}
}

func (p *pricerClient) CalculateTripCost(weightClass, distance int, earnWithFuel bool) int {
	work := p.workToBeDone(weightClass, distance)
	tripCost := p.nominalTripCost(work, p.config.HourlyWage) + p.earnWithRatingPoints(work, p.config.HourlyWage)
	if earnWithFuel {
		return tripCost + p.earnTripFuel(tripCost)
	} else {
		return tripCost
	}
}

func (p *pricerClient) workToBeDone(weightClass, distance int) int {
	return weightClass * int((float64(distance) / math.Pow10(6)))
}

func (p *pricerClient) earnTripFuel(tripCost int) int {
	return p.byminuteWage() * tripCost
}

func (p *pricerClient) nominalTripCost(work, hourlyWage int) int {
	return work * hourlyWage
}

func (p *pricerClient) earnWithRatingPoints(hourlyWage, points int) int { return hourlyWage * points }

// 16% of the total trip cost
func (p *pricerClient) CalculateTripRevenue(tripCost int) int { return (16 / 100) * tripCost }

func (p *pricerClient) byminuteWage() int {
	return p.config.HourlyWage / 60
}
