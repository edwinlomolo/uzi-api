package main

import (
	"fmt"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/gql"
	"github.com/3dw1nM0535/uzi-api/gql/resolvers"
	"github.com/3dw1nM0535/uzi-api/handler"
	"github.com/3dw1nM0535/uzi-api/internal/aws"
	"github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/internal/middleware"
	"github.com/3dw1nM0535/uzi-api/internal/pricer"
	"github.com/3dw1nM0535/uzi-api/internal/route"
	"github.com/3dw1nM0535/uzi-api/services/courier"
	"github.com/3dw1nM0535/uzi-api/services/ipinfo"
	"github.com/3dw1nM0535/uzi-api/services/location"
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

// TODO probably setup server client and user factory paradigm
// to setup its dependencies and services with receiver methods
func main() {
	// Chi router TODO refactor all these to one setup func
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	r.Use(handler.Logger)

	// Services TODO refactor all these to one setup func
	config.LoadConfig()
	cfg := config.GetConfig()
	logger.NewLogger()
	log := logger.GetLogger()
	store.InitializeStorage()
	cache.NewCache(cfg.Database.Redis, log)
	cache.GetCache()
	ipinfo.NewIpinfoService()
	store.GetDatabase()
	user.NewUserService()
	session.NewSessionService()
	courier.NewCourierService()
	aws.NewAwsS3Service()
	upload.NewUploadService()
	location.NewLocationService()
	route.NewRouteService()
	trip.NewTripService()
	pricer.NewPricer()

	// Graphql TODO refactor this to one setup func
	srv := gqlHandler.NewDefaultServer(gql.NewExecutableSchema(resolvers.New()))

	// Routes TODO (look at first route setup comment)
	r.Get("/ipinfo", handler.Ipinfo())
	r.Get("/", playground.Handler("GraphQL playground", "/api"))
	r.Handle("/api", middleware.Auth(srv))
	r.Post("/signin", handler.Signin())
	r.Post("/courier/onboard", handler.CourierOnboarding())
	r.Post("/courier/upload/document", handler.UploadDocument())

	// Server
	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", cfg.Server.Port),
		Handler: r,
	}

	// Run server TODO refactor this to one setup func to start server
	logrus.Fatal(s.ListenAndServe())
}
