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
	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/log"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
	"github.com/pegov/fauth-backend-go/internal/storage"
	"github.com/pegov/fauth-backend-go/internal/token"
)

func main() {
	godotenv.Load()

	logger := log.NewSimpleLogger(log.LevelDebug, true)

	cfg, err := config.New()
	if err != nil {
		logger.Criticalf("Failed to read config, err = %s", err)
	}
	logger.Debugf("CONFIG: \n%s", cfg.Pretty())

	ctx := context.Background()

	db, err := storage.GetDB(
		ctx,
		logger,
		cfg.DatabaseURL,
		cfg.DatabaseMaxIdleConns,
		cfg.DatabaseMaxOpenConns,
		time.Duration(cfg.DatabaseConnMaxLifetime)*time.Second,
	)
	if err != nil {
		logger.Criticalf("Failed to connect to db, err = %s", err)
	}
	defer db.Close()

	cache, err := storage.GetCache(ctx, logger, cfg.CacheURL)
	if err != nil {
		logger.Criticalf("Failed to connect to cache, err = %s", err)
	}
	defer cache.Close()

	userRepo := repo.NewUserRepo(db, cache)

	var captchaClient captcha.CaptchaClient
	if cfg.Debug {
		captchaClient = captcha.NewDebugCaptchaClient("")
	} else {
		captchaClient = captcha.NewReCaptchaClient(cfg.RecaptchaSecret)
	}

	passwordHasher := password.NewBcryptPasswordHasher()
	privateKey, err := os.ReadFile("./id_ed25519_auth_1.key")
	if err != nil {
		logger.Criticalf("Failed to read private key")
	}
	publicKey, err := os.ReadFile("./id_ed25519_auth_1.pub")
	if err != nil {
		logger.Criticalf("Failed to read public key")
	}
	tokenBackend := token.NewJwtBackend(privateKey, publicKey, "1")

	authService := service.NewAuthService(
		userRepo,
		captchaClient,
		passwordHasher,
		tokenBackend,
	)
	adminService := service.NewAdminService(userRepo)

	r, err := api.NewRouter(
		cfg,
		logger,
		authService,
		adminService,
	)
	if err != nil {
		logger.Criticalf("Failed to create router, err = %s", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

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

outer:
	for {
		switch sig := <-signals; sig {
		case syscall.SIGHUP:
			logger.Warnf("Received signal to reset file descriptors for log files: %s", sig)
			if err := logger.Restart(); err != nil {
				logger.Errorf("Failed to reset logger: %s", err)
				break outer
			}
		default:
			logger.Warnf("Received signal: %s", sig)
			break outer
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Criticalf("server.Shutdown() failed, err = %s", err)
	}

	logger.Infof("Bye!")
}
