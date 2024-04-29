package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/pegov/fauth-backend-go/internal/api"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("Failed to load .env file")
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	r := api.NewRouter()
	slog.Info("Starting server...")
	http.ListenAndServe("localhost:3000", r)
}
