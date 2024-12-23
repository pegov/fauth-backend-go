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

func main() {
	godotenv.Load()

	ctx := context.Background()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	cfg, err := config.New()
	checkErr(err, "config.New")
	checkErr(cfg.ParseFlags(os.Args[1:]), "cfg.ParseFlags")

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
	checkErr(err, "api.Prepare")

	checkErr(api.Run(
		ctx,
		cfg,
		logger,
		signals,
		httpServer,
	), "api.Run")
}

func checkErr(err error, description string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: err=%s\n", description, err)
		os.Exit(1)
	}
}
