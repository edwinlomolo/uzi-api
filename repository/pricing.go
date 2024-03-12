package repository

import (
	"context"
	"database/sql"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	sqlStore "github.com/edwinlomolo/uzi-api/store"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PricerRepository struct {
	pricer internal.Pricing
	store  *sqlc.Queries
	log    *logrus.Logger
}

func (p *PricerRepository) Init() {
	p.pricer = internal.GetPricer()
	p.store = sqlStore.GetDb()
	p.log = internal.GetLogger()
}

func (p *PricerRepository) getCourierByID(courierID uuid.UUID) (*model.Courier, error) {
	courier, err := p.store.GetCourierByID(
		context.Background(),
		courierID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		p.log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"error":      err,
		}).Errorf("get courier by id")
		return nil, err
	}

	return &model.Courier{
		ID:        courier.ID,
		ProductID: courier.ProductID.UUID,
	}, nil
}

func (p *PricerRepository) getCourierProduct(productID uuid.UUID) (*model.Product, error) {
	product, err := p.store.GetCourierProductByID(
		context.Background(),
		productID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		p.log.WithFields(logrus.Fields{
			"product_id": productID,
			"error":      err,
		}).Errorf("get courier product")
		return nil, err
	}

	return &model.Product{
		ID:          product.ID,
		Name:        product.Name,
		WeightClass: int(product.WeightClass),
	}, nil
}

func (p *PricerRepository) GetTripCost(trip model.Trip, distance int) (int, error) {
	if trip.CourierID.String() == internal.ZERO_UUID {
		return 0, nil
	}

	courier, err := p.getCourierByID(*trip.CourierID)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"error":    err,
			"distance": distance,
		}).Errorf("get trip courier for trip cost calculation")
		return 0, err
	}

	product, productErr := p.getCourierProduct(courier.ProductID)
	if productErr != nil {
		p.log.WithFields(logrus.Fields{
			"error":              productErr,
			"courier_product_id": courier.ProductID,
		}).Errorf("get trip courier product for trip cost calculation")
		return 0, productErr
	}

	return p.pricer.CalculateTripCost(int(product.WeightClass), distance, product.Name != "UziX"), nil
}
