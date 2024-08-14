package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/pegov/fauth-backend-go/internal/api"
	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/log"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/storage"
)

func main() {
	godotenv.Load()

	logger := log.NewSimpleLogger(log.LevelDebug, true)

	cfg, err := config.New()
	if err != nil {
		logger.Criticalf("Failed to read config, err = %s", err)
	}
	logger.Debugf("CONFIG: \n%s", cfg.Pretty())

	db, err := storage.GetDB(logger, cfg.DatabaseURL)
	if err != nil {
		logger.Criticalf("Failed to connect to db, err = %s", err)
	}
	cache, err := storage.GetCache(logger, cfg.CacheURL)
	if err != nil {
		logger.Criticalf("Failed to connect to cache, err = %s", err)
	}

	userRepo := repo.NewUserRepo(db, cache)

	r, err := api.NewRouter(cfg, logger, userRepo)
	if err != nil {
		logger.Criticalf("Failed to create router, err = %s", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	logger.Infof("Starting server on %s", addr)

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Criticalf("server.ListenAndServe() failed, err = %s", err)
		}
	}()

	logger.Warnf("Received signal: %s", <-signals)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Criticalf("server.Shutdown() failed, err = %s", err)
	}

	logger.Infof("Bye!")
}
