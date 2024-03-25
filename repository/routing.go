package repository

import (
	"context"
	"database/sql"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

type RouteRepository struct {
	store *sqlc.Queries
	log   *logrus.Logger
}

func (r *RouteRepository) Init(q *sqlc.Queries) {
	r.store = q
	r.log = internal.GetLogger()
}

func (r *RouteRepository) GetNearbyAvailableCourierProducts(params sqlc.GetNearbyAvailableCourierProductsParams) ([]*model.Product, error) {
	var nearbyProducts []*model.Product

	nearbys, err := r.store.GetNearbyAvailableCourierProducts(context.Background(), params)
	if err == sql.ErrNoRows {
		return make([]*model.Product, 0), nil
	} else if err != nil {
		r.log.WithFields(logrus.Fields{
			"error": err,
			"args":  params,
		}).Errorf("nearby courier products")
		return nil, err
	}

	for _, item := range nearbys {
		product := &model.Product{
			ID:          item.ID_2,
			Name:        item.Name,
			Description: item.Description,
			IconURL:     item.Icon,
		}

		nearbyProducts = append(nearbyProducts, product)
	}
	return nearbyProducts, nil
}
