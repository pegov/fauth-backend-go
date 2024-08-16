package main

import (
	"context"
	"fmt"
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
		fmt.Printf("[Prepare] %s\n", err)
		os.Exit(1)
	}

	if err := api.Run(
		ctx,
		logger,
		signals,
		httpServer,
	); err != nil {
		fmt.Printf("[Run] %s\n", err)
		os.Exit(1)
	}
}
