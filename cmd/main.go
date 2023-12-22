package main

import (
	"fmt"
	"net/http"

	uzi "github.com/3dw1nM0535/uzi-api"
	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/handler"
	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/pkg/cache"
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

	// Configs
	configs := config.LoadConfig()

	// Logger
	logger := logger.NewLogger()

	// Store
	store, storeErr := store.InitializeStorage(logger, "./store/migrations")
	if storeErr != nil {
		logger.Errorf("%s-%v", "ServerStorageInitializeErr", storeErr.Error())
	}

	// Cache
	cache := cache.NewCache(configs.Database.Redis, logger)

	// Services
	services.NewIpinfoService(cache, configs.Ipinfo, logger)

	// Graphql
	srv := gqlHandler.NewDefaultServer(uzi.NewExecutableSchema(uzi.New(store, cache, logger, configs)))

	// Routes
	r.Handle("/ipinfo", handler.Context(handler.Logger(handler.Ipinfo())))
	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Handle("/query", handler.Context(handler.Logger(srv)))

	// Server
	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", configs.Server.Port),
		Handler: r,
	}

	// Run server
	logrus.Fatal(s.ListenAndServe())
}
