package pricer

import (
	"math"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/sirupsen/logrus"
)

var (
	Pricer Pricing
)

type Pricing interface {
	CalculateTripCost(weightClass, distance int, earnWithFuel bool) int
	CalculateTripRevenue(tripCost int) int
}

type pricerClient struct {
	logger *logrus.Logger
}

func NewPricer() {
	Pricer = &pricerClient{logger.Logger}
}

func (p *pricerClient) CalculateTripCost(
	weightClass, distance int,
	earnWithFuel bool,
) int {
	hourlyWage := config.Config.Pricer.HourlyWage
	work := p.workToBeDone(weightClass, distance)
	tripCost := p.nominalTripCost(work, hourlyWage) + p.earnWithRatingPoints(work, hourlyWage)
	if earnWithFuel {
		return tripCost + p.earnTripFuel(tripCost)
	} else {
		return tripCost
	}
}

func (p *pricerClient) workToBeDone(weightClass, distance int) int {
	return weightClass * distance / int(math.Pow10(6))
}

func (p *pricerClient) earnTripFuel(tripCost int) int {
	return p.byminuteWage() * tripCost
}

func (p *pricerClient) nominalTripCost(work, hourlyWage int) int {
	return work * hourlyWage
}

func (p *pricerClient) earnWithRatingPoints(
	hourlyWage,
	points int,
) int {
	return hourlyWage * points
}

// 16% of the total trip cost
func (p *pricerClient) CalculateTripRevenue(
	tripCost int,
) int {
	return (16 / 100) * tripCost
}

func (p *pricerClient) byminuteWage() int {
	return config.Config.Pricer.HourlyWage / 60
}
