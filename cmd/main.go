package main

import (
	"fmt"
	"net/http"

	uzi "github.com/3dw1nM0535/uzi-api"
	"github.com/3dw1nM0535/uzi-api/aws"
	"github.com/3dw1nM0535/uzi-api/cache"
	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/handler"
	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/middleware"
	"github.com/3dw1nM0535/uzi-api/services"
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
	store.InitializeStorage(logger, "./store/migrations")
	cache := cache.NewCache(configs.Database.Redis, logger)
	services.NewIpinfoService(cache, configs.Ipinfo, logger)
	services.NewUserService(store.GetDatabase(), cache, logger)
	services.NewSessionService(store.GetDatabase(), logger, configs.Jwt)
	services.NewCourierService(logger, store.GetDatabase())
	aws.NewAwsS3Service(configs.Aws, logger)
	services.NewUploadService(aws.GetS3Service(), logger, store.GetDatabase())

	// Graphql
	srv := gqlHandler.NewDefaultServer(uzi.NewExecutableSchema(uzi.New()))

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
