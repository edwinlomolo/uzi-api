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
	userService services.UserService
	redisClient *redis.Client
}

func New() gql.Config {
	c := gql.Config{Resolvers: &Resolver{
		services.GetUploadService(),
		services.GetCourierService(),
		internal.GetLocationService(),
		services.GetTripService(),
		services.GetUserService(),
		internal.GetCache().GetRedis(),
	}}

	return c
}
