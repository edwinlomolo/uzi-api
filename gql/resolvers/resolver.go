package resolvers

import (
	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/courier"
	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/location"
	"github.com/edwinlomolo/uzi-api/routing"
	"github.com/edwinlomolo/uzi-api/trip"
	"github.com/edwinlomolo/uzi-api/upload"
	"github.com/edwinlomolo/uzi-api/user"
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
	routeService routing.Route
	tripService  trip.TripService
	redisClient  *redis.Client
	userService  user.UserService
}

func New() gql.Config {
	c := gql.Config{Resolvers: &Resolver{
		upload.Upload,
		courier.Courier,
		location.Location,
		routing.Routing,
		trip.Trip,
		cache.Redis,
		user.User,
	}}

	return c
}
