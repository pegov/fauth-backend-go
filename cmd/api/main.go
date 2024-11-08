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
	slogger "github.com/pegov/fauth-backend-go/internal/logger"
)

var (
	host                                  string
	port                                  int
	debug, verbose, test                  bool
	accessLog, errorLog                   string
	privateKeyPath, publicKeyPath, jwtKID string
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := cfg.ParseFlags(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var logLevel slog.Level
	if cfg.Flags.Verbose {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slogger.NewColoredHandler(os.Stdout, &slogger.Options{
		Level:    logLevel,
		NoIndent: true,
	}))

	httpServer, err := api.Prepare(
		ctx,
		cfg,
		logger,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api.Prepare err=%s\n", err)
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
		fmt.Fprintf(os.Stderr, "api.Run err=%s\n", err)
		os.Exit(1)
	}
}
