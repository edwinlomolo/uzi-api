package main

import (
	"fmt"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/store"
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

	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", configs.Server.Port),
		Handler: r,
	}

	logrus.Fatal(s.ListenAndServe())
}
