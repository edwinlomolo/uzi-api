package resolvers

import (
	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/internal/cache"
	"github.com/edwinlomolo/uzi-api/internal/route"
	"github.com/edwinlomolo/uzi-api/services/courier"
	"github.com/edwinlomolo/uzi-api/services/location"
	"github.com/edwinlomolo/uzi-api/services/trip"
	"github.com/edwinlomolo/uzi-api/services/upload"
	"github.com/redis/go-redis/v9"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

//go:generate go run github.com/99designs/gqlgen generate --verbose

type Resolver struct {
	upload.UploadService
	courier.CourierService
	location.LocationService
	routeService route.Route
	tripService  trip.TripService
	redisClient  *redis.Client
}

func New() gql.Config {
	c := gql.Config{Resolvers: &Resolver{
		upload.Upload,
		courier.Courier,
		location.Location,
		route.Routing,
		trip.Trip,
		cache.Redis,
	}}

	return c
}
