package resolvers

import (
	"github.com/edwinlomolo/uzi-api/controllers"
	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/redis/go-redis/v9"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

//go:generate go run github.com/99designs/gqlgen generate --verbose

var log = internal.GetLogger()

type Resolver struct {
	controllers.UploadController
	controllers.CourierController
	internal.LocationController
	tripController controllers.TripController
	userController controllers.UserController
	redisClient    *redis.Client
}

func New(q *sqlc.Queries) gql.Config {
	controllers.NewIpinfoController()
	controllers.NewUserController(q)
	controllers.NewUploadController(q)
	controllers.NewCourierController(q)
	controllers.NewTripController(q)
	internal.NewLocationController()

	c := gql.Config{Resolvers: &Resolver{
		controllers.GetUploadController(),
		controllers.GetCourierController(),
		internal.GetLocationController(),
		controllers.GetTripController(),
		controllers.GetUserController(),
		internal.GetCache().GetRedis(),
	}}

	return c
}
