package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/pegov/fauth-backend-go/internal/api"
	"github.com/pegov/fauth-backend-go/internal/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("Failed to load .env file")
	}

	cfg := config.New()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	r := api.NewRouter(&cfg)
	slog.Info("Starting server...")
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("ListenAndServe failed", "err", err)
	}
}
