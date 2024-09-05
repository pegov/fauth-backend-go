package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/pegov/fauth-backend-go/internal/api"
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	httpServer, logger, err := api.Prepare(
		ctx,
		os.Args[1:],
		os.Getenv,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		logger.Error("api.Prepare", slog.Any("err", err))
		os.Exit(1)
	}

	if err := api.Run(
		ctx,
		logger,
		signals,
		httpServer,
	); err != nil {
		logger.Error("api.Run", slog.Any("err", err))
		os.Exit(1)
	}
}
