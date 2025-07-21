package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"

	"github.com/pegov/fauth-backend-go/internal/api"
	"github.com/pegov/fauth-backend-go/internal/config"
	slogger "github.com/pegov/fauth-backend-go/internal/logger"
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var cfg config.Config
	config.CustomizeCLI()
	flags, err := config.ParseFlags(&cfg, config.ParseOptions{RequiredByDefault: true})
	checkErr(err, "parse flags")
	cmd := &cli.Command{
		Name:        "api",
		Usage:       "start api server",
		Description: "fauth backend service",
		Version:     "0.0.1",
		Flags:       flags,
		Action: func(ctx context.Context, c *cli.Command) error {
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

			httpServer, err := api.Prepare(ctx, &cfg, logger)
			if err != nil {
				return fmt.Errorf("api.Prepare: %w", err)
			}

			if err := api.Run(
				ctx,
				&cfg,
				logger,
				signals,
				httpServer,
			); err != nil {
				return fmt.Errorf("api.Run: %w", err)
			}

			return nil
		},
	}

	checkErr(cmd.Run(ctx, os.Args), "cmd.Run")
}

func checkErr(err error, description string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: err=%s\n", description, err)
		os.Exit(1)
	}
}
