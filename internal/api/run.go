package api

import (
	"context"
	"crypto/ed25519"
	"errors"
	"flag"
	"fmt"
	"io"
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
	"github.com/pegov/fauth-backend-go/internal/log"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
	"github.com/pegov/fauth-backend-go/internal/storage"
	"github.com/pegov/fauth-backend-go/internal/token"
)

func Prepare(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdout, stderr io.Writer,
) (*http.Server, log.Logger, error) {
	var (
		host                                  string
		port                                  int
		debug, verbose, trace, test           bool
		accessLog, errorLog                   string
		privateKeyPath, publicKeyPath, jwtKID string
	)

	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.StringVar(&host, "host", "127.0.0.1", "http server host")
	flagSet.IntVar(&port, "port", 15500, "http server port")
	flagSet.BoolVar(&debug, "debug", true, "turn on debug captcha, mail")
	flagSet.BoolVar(&verbose, "verbose", true, "log level = DEBUG")
	flagSet.BoolVar(&trace, "trace", false, "log level = TRACE")
	flagSet.BoolVar(&test, "test", false, "for testing (sqlite, cache in memory)")

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
		return nil, nil, fmt.Errorf("failed to parse args: %w", err)
	}

	var logLevel log.Level
	if trace {
		logLevel = log.LevelTrace
	} else if verbose {
		logLevel = log.LevelDebug
	} else {
		logLevel = log.LevelInfo
	}

	var logger log.Logger
	if accessLog == "" && errorLog == "" {
		logger = log.NewSimpleLogger(logLevel, true)
	} else {
		if accessLog != "" && errorLog != "" {
			var err error
			logger, err = log.NewSeparateLogger(accessLog, errorLog, true)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to setup separate logger: %w", err)
			}
		}
	}

	if trace {
		logger.Warnf("trace flag is ON")
	}

	if verbose {
		logger.Warnf("verbose flag is ON")
	}

	if debug {
		logger.Warnf("debug flag is ON")
	}

	cfg, err := config.New(getenv)
	if err != nil {
		logger.Errorf("Failed to read config")
		return nil, nil, err
	}

	var (
		db    *sqlx.DB
		cache storage.CacheOps
	)
	if test {
		db, err = storage.GetInMemoryDB(ctx, logger, ":memory:")
		if err != nil {
			logger.Errorf("Failed to connect to db: %s", cfg.DatabaseURL)
			return nil, nil, err
		}
		defer db.Close()

		cache = storage.NewMemoryCache()
	} else {
		logger.Debugf("DB URL: %s", cfg.DatabaseURL)
		logger.Debugf("CACHE URL: %s", cfg.CacheURL)

		db, err = storage.GetDB(
			ctx,
			logger,
			cfg.DatabaseURL,
			cfg.DatabaseMaxIdleConns,
			cfg.DatabaseMaxOpenConns,
			time.Duration(cfg.DatabaseConnMaxLifetime)*time.Second,
		)
		if err != nil {
			logger.Errorf("Failed to connect to db: %s", cfg.DatabaseURL)
			return nil, nil, err
		}
		defer db.Close()

		cacheClient, err := storage.GetCache(
			ctx,
			logger,
			cfg.CacheURL,
		)
		if err != nil {
			logger.Errorf("Failed to connect to cache: %s", cfg.CacheURL)
			return nil, nil, err
		}
		defer cacheClient.Close()
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

	var (
		privateKey, publicKey []byte
	)
	if test {
		generateKeys := func(seed []byte) ([]byte, []byte) {
			private := ed25519.NewKeyFromSeed(seed)
			return []byte(private), private.Public().([]byte)
		}
		privateKey, publicKey = generateKeys([]byte(strings.Repeat("a", ed25519.SeedSize)))
		jwtKID = "1"
	} else {
		privateKey, err = os.ReadFile(privateKeyPath)
		if err != nil {
			logger.Errorf("Failed to read private key path: %s", privateKeyPath)
			return nil, nil, err
		}
		publicKey, err = os.ReadFile(publicKeyPath)
		if err != nil {
			logger.Errorf("Failed to read public key path: %s", publicKeyPath)
			return nil, nil, err
		}
		jwtKID = strings.TrimSpace(jwtKID)
		if jwtKID == "" {
			logger.Errorf("jwt kid is empty!")
			return nil, nil, errors.New("jwt kid is empty")
		}
	}
	tokenBackend := token.NewJwtBackend(privateKey, publicKey, jwtKID)

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
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	httpServer := http.Server{
		Addr:    addr,
		Handler: srv,
	}

	return &httpServer, logger, nil
}

func Run(
	ctx context.Context,
	logger log.Logger,
	signals <-chan os.Signal,
	httpServer *http.Server,
) error {
	logger.Infof("Starting http server at %s", httpServer.Addr)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Server failed to listen and serve: %s", err)
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

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Errorf("Failed to gracefully shutdown http server: %s", err)
		return errors.New("failed to do http server gracefull shutdown")
	}

	logger.Infof("Graceful shutdown!")

	return nil
}
