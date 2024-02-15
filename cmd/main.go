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
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql"
	"github.com/edwinlomolo/uzi-api/gql/resolvers"
	"github.com/edwinlomolo/uzi-api/handler"
	"github.com/edwinlomolo/uzi-api/internal/aws"
	"github.com/edwinlomolo/uzi-api/internal/cache"
	"github.com/edwinlomolo/uzi-api/internal/jwt"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/internal/middleware"
	"github.com/edwinlomolo/uzi-api/internal/pricer"
	"github.com/edwinlomolo/uzi-api/internal/route"
	"github.com/edwinlomolo/uzi-api/services/courier"
	"github.com/edwinlomolo/uzi-api/services/ipinfo"
	"github.com/edwinlomolo/uzi-api/services/location"
	"github.com/edwinlomolo/uzi-api/services/session"
	"github.com/edwinlomolo/uzi-api/services/trip"
	"github.com/edwinlomolo/uzi-api/services/upload"
	"github.com/edwinlomolo/uzi-api/services/user"
	"github.com/edwinlomolo/uzi-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

// TODO probably setup server client and user factory paradigm
// to setup its dependencies and services with receiver methods
func main() {
	// Chi router TODO refactor all these to one setup func
	r := chi.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	})
	r.Use(handler.Logger)

	// Services TODO refactor all these to one setup func
	config.LoadConfig()
	logger.NewLogger()
	store.InitializeStorage()
	cache.NewCache(config.Config.Database.Redis, logger.Logger)
	ipinfo.NewIpinfoService()
	jwt.NewJwtService()
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
	srv := gqlHandler.New(gql.NewExecutableSchema(resolvers.New()))
	srv.AddTransport(&transport.POST{})
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

	// Routes TODO (look at first route setup comment)
	r.Get("/ipinfo", handler.Ipinfo())
	r.Get("/", playground.Handler("GraphQL playground", "/api"))
	r.Handle("/api", middleware.Auth(srv))
	r.Post("/signin", handler.Signin())
	r.Post("/user/onboard", handler.UserOnboarding())
	r.Post("/courier/upload/document", handler.UploadDocument())
	r.Handle("/subscription", srv)

	// Server
	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", config.Config.Server.Port),
		Handler: c.Handler(r),
	}

	// Run server TODO refactor this to one setup func to start server
	logrus.Fatal(s.ListenAndServe())
}
