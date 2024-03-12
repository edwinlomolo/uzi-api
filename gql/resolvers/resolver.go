package resolvers

import (
	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/services"
	"github.com/redis/go-redis/v9"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

//go:generate go run github.com/99designs/gqlgen generate --verbose

var log = internal.GetLogger()

type Resolver struct {
	services.UploadService
	services.CourierService
	internal.LocationService
	tripService services.TripService
	redisClient *redis.Client
	userService services.UserService
}

func New() gql.Config {
	redis := internal.GetCache()

	c := gql.Config{Resolvers: &Resolver{
		services.GetUploadService(),
		services.GetCourierService(),
		internal.GetLocationService(),
		services.GetTripService(),
		redis.GetRedis(),
		services.GetUserService(),
	}}

	return c
}
