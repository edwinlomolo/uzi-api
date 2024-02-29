package pricer

import (
	"context"
	"math"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/constants"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/edwinlomolo/uzi-api/store"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

var (
	Pricer Pricing
)

type Pricing interface {
	CalculateTripCost(weightClass, distance int, earnWithFuel bool) int
	GetTripCost(trip model.Trip, distance int) (int, error)
	CalculateTripRevenue(tripCost int) int
}

type pricerClient struct {
	log   *logrus.Logger
	store *sqlc.Queries
}

func NewPricer() {
	Pricer = &pricerClient{logger.Logger, store.DB}
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

func (p *pricerClient) GetTripCost(trip model.Trip, distance int) (int, error) {
	if trip.CourierID.String() == constants.ZERO_UUID {
		return 0, nil
	}

	courier, err := p.store.GetCourierByID(context.Background(), *trip.CourierID)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"error":    err,
			"distance": distance,
		}).Errorf("get trip courier for trip cost calculation")
		return 0, err
	}

	product, productErr := p.store.GetCourierProductByID(context.Background(), courier.ProductID.UUID)
	if productErr != nil {
		p.log.WithFields(logrus.Fields{
			"error":              productErr,
			"courier_product_id": courier.ProductID.UUID,
		}).Errorf("get trip courier product for trip cost calculation")
		return 0, productErr
	}

	return p.CalculateTripCost(int(product.WeightClass), distance, product.Name != "UziX"), nil
}
