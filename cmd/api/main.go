package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/pegov/fauth-backend-go/internal/api"
	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("Failed to load .env file")
	}

	logger := logger.NewSimpleLogger(logger.LevelDebug, true)
	cfg := config.New(logger)

	r := api.NewRouter(&cfg, logger)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	logger.Infof("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Criticalf("ListenAndServe failed, err = %s", err)
	}
}
