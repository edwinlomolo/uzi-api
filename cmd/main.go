package main

import (
	"fmt"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/gql"
	"github.com/3dw1nM0535/uzi-api/gql/resolvers"
	"github.com/3dw1nM0535/uzi-api/handler"
	"github.com/3dw1nM0535/uzi-api/pkg/aws"
	"github.com/3dw1nM0535/uzi-api/pkg/cache"
	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	"github.com/3dw1nM0535/uzi-api/pkg/middleware"
	"github.com/3dw1nM0535/uzi-api/services/courier"
	"github.com/3dw1nM0535/uzi-api/services/ipinfo"
	"github.com/3dw1nM0535/uzi-api/services/location"
	"github.com/3dw1nM0535/uzi-api/services/route"
	"github.com/3dw1nM0535/uzi-api/services/session"
	"github.com/3dw1nM0535/uzi-api/services/trip"
	"github.com/3dw1nM0535/uzi-api/services/upload"
	"github.com/3dw1nM0535/uzi-api/services/user"
	"github.com/3dw1nM0535/uzi-api/store"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	// Chi router
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	r.Use(handler.Logger)

	// Services
	configs := config.LoadConfig()
	logger := logger.NewLogger()
	store.InitializeStorage(logger, configs.Database.Rdbms.MigrationUrl)
	cache := cache.NewCache(configs.Database.Redis, logger)
	ipinfo.NewIpinfoService(cache, configs.Ipinfo, logger)
	user.NewUserService(store.GetDatabase(), cache, logger)
	session.NewSessionService(store.GetDatabase(), logger, configs.Jwt)
	courier.NewCourierService(logger, store.GetDatabase())
	aws.NewAwsS3Service(configs.Aws, logger)
	upload.NewUploadService(aws.GetS3Service(), logger, store.GetDatabase())
	location.NewLocationService(configs.GoogleMaps, logger, cache)
	trip.NewTripService(logger, store.GetDatabase())
	route.NewRouteService(cache, logger, store.GetDatabase(), configs.GoogleMaps)

	// Graphql
	srv := gqlHandler.NewDefaultServer(gql.NewExecutableSchema(resolvers.New()))

	// Routes
	r.Get("/ipinfo", handler.Ipinfo())
	r.Get("/", playground.Handler("GraphQL playground", "/api"))
	r.Handle("/api", middleware.Auth(srv))
	r.Post("/signin", handler.Signin())
	r.Post("/courier/onboard", handler.CourierOnboarding())
	r.Post("/courier/upload/document", handler.UploadDocument())

	// Server
	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", configs.Server.Port),
		Handler: r,
	}

	// Run server
	logrus.Fatal(s.ListenAndServe())
}
