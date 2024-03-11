package main

import (
	"fmt"
	"net/http"
	"time"

	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/gql/resolvers"
	"github.com/edwinlomolo/uzi-api/handler"
	"github.com/edwinlomolo/uzi-api/ipinfo"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/edwinlomolo/uzi-api/middleware"
	"github.com/edwinlomolo/uzi-api/store"
	"github.com/edwinlomolo/uzi-api/uploader"
	"github.com/edwinlomolo/uzi-api/user"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// TODO I can set this up better
func main() {
	// Config
	config.LoadConfig()

	// Logger
	log := logger.New()

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.Config.Sentry.Dsn,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		log.WithError(err).Errorf("sentry http middleware")
	}

	// Server Routing
	r := chi.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	})
	// Middleware
	r.Use(middleware.GetIp)
	r.Use(middleware.Sentry)
	r.Use(middleware.Logger)

	// Database queries
	queries, _ := store.InitializeStorage()

	// Redis cache client
	cache := cache.New()

	// Services
	ipinfoService := ipinfo.New(cache)
	userService := user.New(queries, cache)
	uploader.New()

	srv := gqlHandler.New(gql.NewExecutableSchema(resolvers.New(queries, cache, userService)))
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.SetQueryCache(lru.New(1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(1000),
	})

	r.Route("/v1", func(r chi.Router) {
		r.With(middleware.Auth).Handle("/api", srv)
		r.Post("/signin", handler.Signin(userService))
		r.Post("/user/onboard", handler.UserOnboarding(userService))
		r.Post("/courier/upload/document", handler.UploadDocument())
		r.Get("/ipinfo", handler.Ipinfo(ipinfoService))
	})
	r.Get("/", playground.Handler("GraphQL playground", "/v1/api"))
	r.Handle("/subscription", srv)

	// Server
	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", config.Config.Server.Port),
		Handler: c.Handler(r),
	}

	log.Fatalln(s.ListenAndServe())
}
