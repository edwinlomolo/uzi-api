package main

import (
	"context"
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
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)

	// Logging
	logger := logger.NewLogger()

	// configs
	config.LoadConfig()
	configs := config.GetConfig()

	// Storage
	_, err := store.InitializeStorage(logger, "./store/migrations")
	if err != nil {
		logger.Fatalln(err)
	}

	// New context
	ctx := context.Background()

	// Cache
	cache := cache.NewCache(configs.Database.Redis, logger)

	// Services
	ipinfoService := services.NewIpinfoService(cache, configs.Ipinfo, logger)

	// App context
	ctx = context.WithValue(ctx, "logger", logger)
	ctx = context.WithValue(ctx, "ipinfoService", ipinfoService)

	// Graphql
	srv := gqlHandler.NewDefaultServer(uzi.NewExecutableSchema(uzi.New()))

	r.Handle("/ipinfo", handler.Context(ctx, handler.Logger(handler.Ipinfo())))
	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Handle("/query", handler.Context(ctx, handler.Logger(srv)))

	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", configs.Server.Port),
		Handler: r,
	}

	logrus.Fatal(s.ListenAndServe())
}
