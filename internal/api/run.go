package api

import (
	"context"
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

	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/log"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
	"github.com/pegov/fauth-backend-go/internal/storage"
	"github.com/pegov/fauth-backend-go/internal/token"
)

func Run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdout, stderr io.Writer,
	signals <-chan os.Signal,
) error {
	var (
		host                                  string
		port                                  int
		debug, debugPasswordHasher, verbose   bool
		accessLog, errorLog                   string
		privateKeyPath, publicKeyPath, jwtKID string
	)

	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.StringVar(&host, "host", "127.0.0.1", "http server host")
	flagSet.IntVar(&port, "port", 15500, "http server port")
	flagSet.BoolVar(&debug, "debug", true, "turn on debug captcha, mail")
	flagSet.BoolVar(
		&debugPasswordHasher,
		"debug-password-hasher",
		false,
		"turn on debug password hasher",
	)
	flagSet.BoolVar(&verbose, "verbose", true, "log level = DEBUG")

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
		return fmt.Errorf("failed to parse args: %w", err)
	}

	var logLevel log.Level
	if verbose {
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
				return fmt.Errorf("failed to setup separate logger: %w", err)
			}
		}
	}

	if verbose {
		logger.Warnf("verbose flag is ON")
	}

	if debug {
		logger.Warnf("debug flag is ON")
	}

	if debugPasswordHasher {
		logger.Warnf("debug-password-hasher flag is ON")
	}

	cfg, err := config.New(getenv)
	if err != nil {
		logger.Errorf("Failed to read config")
		return err
	}

	logger.Debugf("DB URL: %s", cfg.DatabaseURL)
	logger.Debugf("CACHE URL: %s", cfg.CacheURL)

	db, err := storage.GetDB(
		ctx,
		logger,
		cfg.DatabaseURL,
		cfg.DatabaseMaxIdleConns,
		cfg.DatabaseMaxOpenConns,
		time.Duration(cfg.DatabaseConnMaxLifetime)*time.Second,
	)
	if err != nil {
		logger.Errorf("Failed to connect to db: %s", cfg.DatabaseURL)
		return err
	}
	defer db.Close()

	cache, err := storage.GetCache(
		ctx,
		logger,
		cfg.CacheURL,
	)
	if err != nil {
		logger.Errorf("Failed to connect to cache: %s", cfg.CacheURL)
		return err
	}
	defer cache.Close()

	userRepo := repo.NewUserRepo(db, cache)

	var captchaClient captcha.CaptchaClient
	if debug {
		captchaClient = captcha.NewDebugCaptchaClient("")
	} else {
		captchaClient = captcha.NewReCaptchaClient(cfg.RecaptchaSecret)
	}

	var passwordHasher password.PasswordHasher
	if debugPasswordHasher {
		passwordHasher = password.NewPlainTextPasswordHasher()
	} else {
		passwordHasher = password.NewBcryptPasswordHasher()
	}

	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		logger.Errorf("Failed to read private key path: %s", privateKeyPath)
		return err
	}
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		logger.Errorf("Failed to read public key path: %s", publicKeyPath)
		return err
	}
	jwtKID = strings.TrimSpace(jwtKID)
	if jwtKID == "" {
		logger.Errorf("jwt kid is empty!")
		return errors.New("jwt kid is empty")
	}
	tokenBackend := token.NewJwtBackend(privateKey, publicKey, jwtKID)

	authService := service.NewAuthService(
		userRepo,
		captchaClient,
		passwordHasher,
		tokenBackend,
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

	logger.Infof("Starting http server at %s", addr)
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
