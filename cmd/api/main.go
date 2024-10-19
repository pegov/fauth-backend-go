package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/pegov/fauth-backend-go/internal/api"
	"github.com/pegov/fauth-backend-go/internal/config"
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	cfg, err := config.New()
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to read config")
		return
	}

	httpServer, logger, host, port, err := api.Prepare(
		ctx,
		cfg,
		os.Args[1:],
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
		host,
		port,
	); err != nil {
		logger.Error("api.Run", slog.Any("err", err))
		os.Exit(1)
	}
}
