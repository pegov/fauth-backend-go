package api

import (
	"context"
	"crypto/ed25519"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/email"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
	"github.com/pegov/fauth-backend-go/internal/storage"
	"github.com/pegov/fauth-backend-go/internal/token"
)

func Prepare(
	ctx context.Context,
	cfg *config.Config,
	args []string,
	stdout, stderr io.Writer,
) (http.Handler, *slog.Logger, string, int, error) {
	var (
		host                                  string
		port                                  int
		debug, verbose, test                  bool
		accessLog, errorLog                   string
		privateKeyPath, publicKeyPath, jwtKID string
	)

	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.StringVar(&host, "host", "127.0.0.1", "http server host")
	flagSet.IntVar(&port, "port", 15500, "http server port")
	flagSet.BoolVar(&debug, "debug", true, "turn on debug captcha, mail")
	flagSet.BoolVar(&verbose, "verbose", true, "log level = DEBUG")
	flagSet.BoolVar(&test, "test", false, "for testing (cache in memory)")

	flagSet.StringVar(&accessLog, "access-log", "", "path to access log file")
	flagSet.StringVar(&errorLog, "error-log", "", "path to error log file")

	flagSet.StringVar(
		&privateKeyPath,
		"jwt-private-key",
		"",
		"path to private key file",
	)
	flagSet.StringVar(
		&publicKeyPath,
		"jwt-public-key",
		"",
		"path to public key file",
	)
	flagSet.StringVar(
		&jwtKID,
		"jwt-kid",
		"",
		"path to public key file",
	)

	if err := flagSet.Parse(args); err != nil {
		return nil, nil, "", 0, fmt.Errorf("failed to parse args: %w", err)
	}

	var logLevel slog.Level
	if verbose {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	if verbose {
		logger.Warn("verbose flag is ON")
	}

	if debug {
		logger.Warn("debug flag is ON")
	}

	var (
		db    *sqlx.DB
		cache storage.CacheOps
		err   error
	)
	if test {
		db, err = storage.GetDB(
			ctx,
			logger,
			cfg.DatabaseURL,
			cfg.DatabaseMaxIdleConns,
			cfg.DatabaseMaxOpenConns,
			time.Duration(cfg.DatabaseConnMaxLifetime)*time.Second,
		)
		if err != nil {
			logger.Error("Failed to connect to db", slog.String("db", cfg.DatabaseURL))
			return nil, nil, "", 0, err
		}

		cache = storage.NewMemoryCache()
	} else {
		logger.Debug("", slog.String("db", cfg.DatabaseURL))
		logger.Debug("", slog.String("cache", cfg.CacheURL))

		db, err = storage.GetDB(
			ctx,
			logger,
			cfg.DatabaseURL,
			cfg.DatabaseMaxIdleConns,
			cfg.DatabaseMaxOpenConns,
			time.Duration(cfg.DatabaseConnMaxLifetime)*time.Second,
		)
		if err != nil {
			logger.Error("Failed to connect to db", slog.String("db", cfg.DatabaseURL))
			return nil, nil, "", 0, err
		}

		cacheClient, err := storage.GetCache(
			ctx,
			logger,
			cfg.CacheURL,
		)
		if err != nil {
			logger.Error("Failed to connect to cache", slog.String("cache", cfg.CacheURL))
			return nil, nil, "", 0, err
		}
		cache = storage.NewRedisCacheWrapper(cacheClient)
	}

	userRepo := repo.NewUserRepo(db, cache)

	var captchaClient captcha.CaptchaClient
	if debug {
		captchaClient = captcha.NewDebugCaptchaClient("")
	} else {
		captchaClient = captcha.NewReCaptchaClient(cfg.RecaptchaSecret)
	}

	var passwordHasher password.PasswordHasher
	if test {
		passwordHasher = password.NewPlainTextPasswordHasher()
	} else {
		passwordHasher = password.NewBcryptPasswordHasher()
	}

	var tokenBackend token.JwtBackend
	if test {
		generateKeys := func(seed []byte) ([]byte, []byte) {
			private := ed25519.NewKeyFromSeed(seed)
			public := private.Public().(ed25519.PublicKey)
			return private, public
		}
		privateKey, publicKey := generateKeys([]byte(strings.Repeat("a", ed25519.SeedSize)))
		jwtKID = "1"
		tokenBackend = token.NewJwtBackendRaw(privateKey, publicKey, jwtKID)
	} else {
		privateKey, err := os.ReadFile(privateKeyPath)
		if err != nil {
			logger.Error(
				"Failed to read private key path",
				slog.String("privateKeyPath", privateKeyPath),
			)
			return nil, nil, "", 0, err
		}
		publicKey, err := os.ReadFile(publicKeyPath)
		if err != nil {
			logger.Error(
				"Failed to read public key path",
				slog.String("publicKeyPath", publicKeyPath),
			)
			return nil, nil, "", 0, err
		}
		jwtKID = strings.TrimSpace(jwtKID)
		if jwtKID == "" {
			logger.Error("jwt kid is empty!")
			return nil, nil, "", 0, errors.New("jwt kid is empty")
		}
		tokenBackend = token.NewJwtBackend(privateKey, publicKey, jwtKID)
	}

	emailClient := email.NewStdEmailClient(
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPHost,
		cfg.SMTPPort,
	)

	authService := service.NewAuthService(
		userRepo,
		captchaClient,
		passwordHasher,
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

	return srv, logger, host, port, nil
}

func Run(
	ctx context.Context,
	logger *slog.Logger,
	signals <-chan os.Signal,
	handler http.Handler,
	host string,
	port int,
) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
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
