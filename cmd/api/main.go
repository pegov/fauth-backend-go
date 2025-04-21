package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
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

	cfg := &config.Config{}

	helpPrinterCustomOrig := cli.HelpPrinterCustom
	cli.HelpPrinterCustom = func(out io.Writer, templ string, data any, customFuncs map[string]any) {
		if customFuncs == nil {
			customFuncs = map[string]any{}
		}
		customFuncs["isEmpty"] = func(v any) bool {
			return reflect.ValueOf(v).IsZero()
		}
		helpPrinterCustomOrig(out, templ, data, customFuncs)
	}

	cli.RootCommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}[global options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}
{{- if len .Authors}}

AUTHOR{{template "authorsTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

GLOBAL OPTIONS:{{range .VisibleFlagCategories}}
   {{if .Name}}{{.Name}}

   {{end}}{{$flglen := len .Flags}}{{range $i, $e := .Flags}}{{if eq (subtract $flglen $i) 1}}{{$e}} {{if $e.Required}}REQUIRED{{else}}{{if and (isEmpty $e.Value) (not $e.IsBoolFlag)}}OPTIONAL{{end}}{{end}}
{{else}}{{$e}} {{if $e.Required}}REQUIRED{{else}}{{if and (isEmpty $e.Value) (not $e.IsBoolFlag)}}OPTIONAL{{end}}{{end}}{{end}}
   {{end}}{{end}}{{else if .VisibleFlags}}

GLOBAL OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{template "copyrightTemplate" .}}{{end}}
`

	cmd := &cli.Command{
		Name:        "api",
		Usage:       "fauth backend service",
		Description: "start api server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Category:    "Database",
				Name:        "db-url",
				Destination: &cfg.Database.URL,
				Sources:     cli.EnvVars("DATABASE_URL"),
				Local:       true,
				Required:    true,
			},
			&cli.IntFlag{
				Category:    "Database",
				Name:        "db-max-idle-conns",
				Destination: &cfg.Database.MaxIdleConns,
				Sources:     cli.EnvVars("DATABASE_MAX_IDLE_CONNS"),
				Local:       true,
				Value:       20,
			},
			&cli.IntFlag{
				Category:    "Database",
				Name:        "db-max-open-conns",
				Destination: &cfg.Database.MaxOpenConns,
				Sources:     cli.EnvVars("DATABASE_MAX_OPEN_CONNS"),
				Local:       true,
				Value:       20,
			},
			&cli.IntFlag{
				Category:    "Database",
				Name:        "db-conn-max-lifetime",
				Destination: &cfg.Database.ConnMaxLifetime,
				Sources:     cli.EnvVars("DATABASE_CONN_MAX_LIFETIME"),
				Local:       true,
				Value:       600,
			},
			&cli.StringFlag{
				Category:    "Cache",
				Name:        "cache-url",
				Destination: &cfg.Cache.URL,
				Sources:     cli.EnvVars("CACHE_URL"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "HTTP",
				Name:        "http-domain",
				Destination: &cfg.HTTP.Domain,
				Sources:     cli.EnvVars("HTTP_DOMAIN"),
				Local:       true,
				Required:    true,
			},
			&cli.BoolFlag{
				Category:    "HTTP",
				Name:        "http-secure",
				Destination: &cfg.HTTP.Secure,
				Sources:     cli.EnvVars("HTTP_SECURE"),
				Local:       true,
				HideDefault: true,
			},
			&cli.StringFlag{
				Category:    "SMTP",
				Name:        "smtp-username",
				Destination: &cfg.SMTP.Username,
				Sources:     cli.EnvVars("SMTP_USERNAME"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "SMTP",
				Name:        "smtp-password",
				Destination: &cfg.SMTP.Password,
				Sources:     cli.EnvVars("SMTP_PASSWORD"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "SMTP",
				Name:        "smtp-host",
				Destination: &cfg.SMTP.Host,
				Sources:     cli.EnvVars("SMTP_HOST"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "SMTP",
				Name:        "smtp-port",
				Destination: &cfg.SMTP.Port,
				Sources:     cli.EnvVars("SMTP_PORT"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "reCAPTCHA",
				Name:        "recaptcha-secret",
				Destination: &cfg.Captcha.RecaptchaSecret,
				Sources:     cli.EnvVars("RECAPTCHA_SECRET"),
				Local:       true,
				Required:    false,
			},
			&cli.StringSliceFlag{
				Category:    "OAuth",
				Name:        "oauth-providers",
				Destination: &cfg.OAuth.Providers,
				Sources:     cli.EnvVars("OAUTH_PROVIDERS"),
				Local:       true,
			},
			&cli.StringFlag{
				Category:    "OAuth",
				Name:        "oauth-google-client-id",
				Destination: &cfg.OAuth.GoogleClientID,
				Sources:     cli.EnvVars("OAUTH_GOOGLE_CLIENT_ID"),
				Local:       true,
			},
			&cli.StringFlag{
				Category:    "OAuth",
				Name:        "oauth-google-client-secret",
				Destination: &cfg.OAuth.GoogleClientSecret,
				Sources:     cli.EnvVars("OAUTH_GOOGLE_CLIENT_SECRET"),
				Local:       true,
			},
			&cli.StringFlag{
				Category:    "OAuth",
				Name:        "oauth-vk-app-id",
				Destination: &cfg.OAuth.VKAppID,
				Sources:     cli.EnvVars("OAUTH_VK_APP_ID"),
				Local:       true,
			},
			&cli.StringFlag{
				Category:    "OAuth",
				Name:        "oauth-vk-app-secret",
				Destination: &cfg.OAuth.VKAppSecret,
				Sources:     cli.EnvVars("OAUTH_VK_APP_SECRET"),
				Local:       true,
			},
			&cli.IntFlag{
				Category:    "App",
				Name:        "login-ratelimit",
				Destination: &cfg.App.LoginRatelimit,
				Sources:     cli.EnvVars("APP_LOGIN_RATELIMIT"),
				Local:       true,
				Value:       10,
			},
			&cli.IntFlag{
				Category:    "App",
				Name:        "access-token-expiration",
				Destination: &cfg.App.AcessTokenExpiration,
				Sources:     cli.EnvVars("APP_ACCESS_TOKEN_EXPIRATION"),
				Local:       true,
				Value:       60 * 60 * 6,
			},
			&cli.IntFlag{
				Category:    "App",
				Name:        "refresh-token-expiration",
				Destination: &cfg.App.RefreshTokenExpiration,
				Sources:     cli.EnvVars("APP_REFRESH_TOKEN_EXPIRATION"),
				Local:       true,
				Value:       60 * 60 * 24 * 30,
			},
			&cli.StringFlag{
				Category:    "App",
				Name:        "access-token-cookie-name",
				Destination: &cfg.App.AccessTokenCookieName,
				Sources:     cli.EnvVars("APP_ACCESS_TOKEN_COOKIE_NAME"),
				Local:       true,
				Value:       "access",
			},
			&cli.StringFlag{
				Category:    "App",
				Name:        "refresh-token-cookie-name",
				Destination: &cfg.App.RefreshTokenCookieName,
				Sources:     cli.EnvVars("APP_REFRESH_TOKEN_COOKIE_NAME"),
				Local:       true,
				Value:       "refresh",
			},
			&cli.StringFlag{
				Category:    "Flags",
				Name:        "host",
				Destination: &cfg.Flags.Host,
				Sources:     cli.EnvVars("HOST"),
				Local:       true,
				Required:    true,
			},
			&cli.IntFlag{
				Category:    "Flags",
				Name:        "port",
				Destination: &cfg.Flags.Port,
				Sources:     cli.EnvVars("PORT"),
				Local:       true,
				Required:    true,
			},
			&cli.BoolFlag{
				Category:    "Flags",
				Name:        "debug",
				Destination: &cfg.Flags.Debug,
				Sources:     cli.EnvVars("DEBUG"),
				Local:       true,
				HideDefault: true,
			},
			&cli.BoolFlag{
				Category:    "Flags",
				Name:        "verbose",
				Destination: &cfg.Flags.Verbose,
				Sources:     cli.EnvVars("VERBOSE"),
				Local:       true,
				HideDefault: true,
			},
			&cli.BoolFlag{
				Category:    "Flags",
				Name:        "test",
				Destination: &cfg.Flags.Test,
				Sources:     cli.EnvVars("TEST"),
				Local:       true,
				HideDefault: true,
			},
			&cli.StringFlag{
				Category:    "Flags",
				Name:        "access-log",
				Destination: &cfg.Flags.AccessLog,
				Sources:     cli.EnvVars("ACCESS_LOG"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "Flags",
				Name:        "error-log",
				Destination: &cfg.Flags.ErrorLog,
				Sources:     cli.EnvVars("ERROR_LOG"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "Flags",
				Name:        "jwt-private-key-path",
				Destination: &cfg.Flags.PrivateKeyPath,
				Sources:     cli.EnvVars("JWT_PRIVATE_KEY_PATH"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "Flags",
				Name:        "jwt-public-key-path",
				Destination: &cfg.Flags.PrivateKeyPath,
				Sources:     cli.EnvVars("JWT_PUBLIC_KEY_PATH"),
				Local:       true,
				Required:    true,
			},
			&cli.StringFlag{
				Category:    "Flags",
				Name:        "jwt-kid",
				Destination: &cfg.Flags.JWTKID,
				Sources:     cli.EnvVars("JWT_KID"),
				Local:       true,
				Required:    true,
			},
		},
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

			httpServer, err := api.Prepare(
				ctx,
				cfg,
				logger,
				os.Stdout,
				os.Stderr,
			)
			if err != nil {
				return fmt.Errorf("api.Prepare: %w", err)
			}

			if err := api.Run(
				ctx,
				cfg,
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
