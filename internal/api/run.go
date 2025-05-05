package api

import (
	"context"
	"crypto/ed25519"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/email"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
	"github.com/pegov/fauth-backend-go/internal/storage"
	"github.com/pegov/fauth-backend-go/internal/token"
)

func PrepareForTest(
	ctx context.Context,
	cfg *config.Config,
	logger *slog.Logger,
	stdout, stderr io.Writer,
) (http.Handler, error) {
	db, err := storage.GetDB(
		ctx,
		logger,
		cfg.Database.URL,
		cfg.Database.MaxIdleConns,
		cfg.Database.MaxOpenConns,
		time.Duration(cfg.Database.ConnMaxLifetime)*time.Second,
	)
	if err != nil {
		logger.Error("Failed to connect to db", slog.String("db", cfg.Database.URL))
		return nil, err
	}

	cache := storage.NewMemoryCache()

	userRepo := repo.NewUserRepo(db, cache)

	passwordManager := password.NewPlainTextPasswordHasher()
	captchaClient := captcha.NewDebugCaptchaClient("")

	generateKeys := func(seed []byte) ([]byte, []byte) {
		private := ed25519.NewKeyFromSeed(seed)
		public := private.Public().(ed25519.PublicKey)
		return private, public
	}
	privateKey, publicKey := generateKeys([]byte(strings.Repeat("a", ed25519.SeedSize)))
	cfg.Flags.JWTKID = "1"
	tokenBackend := token.NewJwtBackendRaw(privateKey, publicKey, cfg.Flags.JWTKID)

	emailClient := email.NewMockEmailClient()

	authService := service.NewAuthService(
		userRepo,
		captchaClient,
		passwordManager,
		tokenBackend,
		emailClient,
	)

	adminService := service.NewAdminService(userRepo)

	srv := NewServer(
		cfg,
		logger,
		authService,
		adminService,
	)

	return srv, nil
}

func Prepare(
	ctx context.Context,
	cfg *config.Config,
	logger *slog.Logger,
	stdout, stderr io.Writer,
) (http.Handler, error) {
	if cfg.Flags.Verbose {
		logger.Warn("verbose flag is ON")
	}

	if cfg.Flags.Debug {
		logger.Warn("debug flag is ON")
	}

	var (
		db    storage.DB
		cache storage.CacheOps
		err   error
	)
	logger.Debug("", slog.String("db", cfg.Database.URL))
	logger.Debug("", slog.String("cache", cfg.Cache.URL))

	db, err = storage.GetDB(
		ctx,
		logger,
		cfg.Database.URL,
		cfg.Database.MaxIdleConns,
		cfg.Database.MaxOpenConns,
		time.Duration(cfg.Database.ConnMaxLifetime)*time.Second,
	)
	if err != nil {
		logger.Error("Failed to connect to db", slog.String("db", cfg.Database.URL), slog.Any("err", err))
		return nil, err
	}

	cacheClient, err := storage.GetCache(
		ctx,
		logger,
		cfg.Cache.URL,
	)
	if err != nil {
		logger.Error("Failed to connect to cache", slog.String("cache", cfg.Cache.URL))
		return nil, err
	}
	cache = storage.NewRedisCacheWrapper(cacheClient)

	userRepo := repo.NewUserRepo(db, cache)

	var captchaClient captcha.CaptchaClient
	if cfg.Flags.Debug {
		captchaClient = captcha.NewDebugCaptchaClient("")
	} else {
		captchaClient = captcha.NewReCaptchaClient(cfg.Captcha.RecaptchaSecret)
	}

	passwordManager := password.NewBcryptPasswordHasher()

	privateKey, err := os.ReadFile(cfg.Flags.PrivateKeyPath)
	if err != nil {
		logger.Error(
			"Failed to read private key path",
			slog.String("privateKeyPath", cfg.Flags.PrivateKeyPath),
		)
		return nil, err
	}
	publicKey, err := os.ReadFile(cfg.Flags.PublicKeyPath)
	if err != nil {
		logger.Error(
			"Failed to read public key path",
			slog.String("publicKeyPath", cfg.Flags.PublicKeyPath),
		)
		return nil, err
	}
	cfg.Flags.JWTKID = strings.TrimSpace(cfg.Flags.JWTKID)
	if cfg.Flags.JWTKID == "" {
		logger.Error("jwt kid is empty!")
		return nil, errors.New("jwt kid is empty")
	}
	tokenBackend := token.NewJwtBackend(privateKey, publicKey, cfg.Flags.JWTKID)

	emailClient := email.NewStdEmailClient(
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
	)

	authService := service.NewAuthService(
		userRepo,
		captchaClient,
		passwordManager,
		tokenBackend,
		emailClient,
	)

	adminService := service.NewAdminService(userRepo)

	srv := NewServer(
		cfg,
		logger,
		authService,
		adminService,
	)

	return srv, nil
}

func Run(
	ctx context.Context,
	cfg *config.Config,
	logger *slog.Logger,
	signals <-chan os.Signal,
	handler http.Handler,
) error {
	addr := net.JoinHostPort(cfg.Flags.Host, strconv.Itoa(cfg.Flags.Port))
	httpServer := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	logger.Info("Starting http server", slog.String("addr", httpServer.Addr))
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to listen and serve", slog.Any("err", err))
		}
	}()

outer:
	for {
		select {
		case sig := <-signals:
			switch sig {
			case syscall.SIGHUP:
				logger.Warn(
					"Received signal to reset file descriptors for log files",
					slog.Any("sig", sig),
				)
			case syscall.SIGTERM, syscall.SIGINT:
				logger.Warn("Received exit signal", slog.Any("sig", sig))
				break outer
			}
		case <-ctx.Done():
			logger.Warn("Context was canceled")
			break outer
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Failed to gracefully shutdown http server", slog.Any("err", err))
		return errors.New("failed to do http server gracefull shutdown")
	}

	logger.Info("Graceful shutdown!")

	return nil
}
