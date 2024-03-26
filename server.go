package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/gql/resolvers"
	"github.com/edwinlomolo/uzi-api/handler"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/middleware"
	"github.com/edwinlomolo/uzi-api/store"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// TODO I can set this up better
func main() {
	// Config
	config.LoadConfig()
	// Logger
	log := internal.NewLogger()
	// Database queries
	q, _ := store.InitializeStorage()
	// Redis cache client
	internal.NewCache()

	isProd := func() bool {
		return config.Config.Server.Env == "production" ||
			config.Config.Server.Env == "staging"
	}

	// Sentry logging setup
	if isProd() {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              config.Config.Sentry.Dsn,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
		}); err != nil {
			log.WithError(err).Errorf("sentry http middleware")
		}
	}

	// Internal services
	internal.NewPricer()
	internal.NewUploader()

	srv := gqlHandler.New(gql.NewExecutableSchema(resolvers.New(q)))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		// TODO testing graph subscription with api clients I'm getting abnormalclosure error???
		ErrorFunc: func(ctx context.Context, err error) {
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				log.WithError(err).Errorf("websocket")
			}
		},
	})
	srv.SetQueryCache(lru.New(1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(1000),
	})

	// Server Routing
	r := chi.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
	})
	// Middleware
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.EnrichRequestContext)
	r.Use(middleware.Sentry)
	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.With(middleware.Auth).Handle("/graphql", srv)
		r.Post("/signin", handler.Signin())
		r.Post("/user/onboard", handler.UserOnboarding())
		r.Post("/courier/upload/document", handler.UploadDocument())
		r.Get("/ipinfo", handler.Ipinfo())
		r.Post("/account/delete", handler.SoftDeleteAccount())
	})
	r.Get("/", playground.Handler("GraphQL playground", "/api/graphql"))
	r.Handle("/subscription", srv)

	// Server
	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", config.Config.Server.Port),
		Handler: c.Handler(r),
	}

	log.Fatalln(s.ListenAndServe())
}
