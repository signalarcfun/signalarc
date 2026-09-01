package main

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/wahyu241205/SignalArc/backend/internal/api"
	"github.com/wahyu241205/SignalArc/backend/internal/config"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

func main() {
	cfg := config.Load()
	if err := cfg.ValidateDatabaseURL(); err != nil {
		log.Fatal().Err(err).Msg("invalid backend config")
	}

	db, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("database startup check failed")
	}
	defer db.Close()

	router := api.NewRouter(db)

	log.Info().Str("env", cfg.AppEnv).Str("port", cfg.AppPort).Msg("starting SignalArc API")

	if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
