package config

import (
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Database ConfigDatabase
	Cache    ConfigCache
	HTTP     ConfigHTTP
	SMTP     ConfigSMTP
	Captcha  ConfigCaptcha
	OAuth    ConfigOAuth
	App      ConfigApp
	Flags    Flags
}

type ConfigDatabase struct {
	URL             string `envconfig:"DATABASE_URL"`
	MaxIdleConns    int    `envconfig:"DATABASE_MAX_IDLE_CONNS"`
	MaxOpenConns    int    `envconfig:"DATABASE_MAX_OPEN_CONNS"`
	ConnMaxLifetime int    `envconfig:"DATABASE_CONN_MAX_LIFETIME"`
}

type ConfigCache struct {
	URL string `envconfig:"CACHE_URL"`
}

type ConfigHTTP struct {
	Domain string `envconfig:"HTTP_DOMAIN"`
	Secure bool   `envconfig:"HTTP_SECURE"`
}

type ConfigSMTP struct {
	Username string `envconfig:"SMTP_USERNAME"`
	Password string `envconfig:"SMTP_PASSWORD"`
	Host     string `envconfig:"SMTP_HOST"`
	Port     string `envconfig:"SMTP_PORT"`
}

type ConfigCaptcha struct {
	RecaptchaSecret string `envconfig:"RECAPTCHA_SECRET"`
}

type ConfigOAuth struct {
	SocialProviders    []string `envconfig:"SOCIAL_PROVIDERS"`
	GoogleClientID     string   `envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string   `envconfig:"GOOGLE_CLIENT_SECRET"`
	VKAppID            string   `envconfig:"VK_APP_ID"`
	VKAppSecret        string   `envconfig:"VK_APP_SECRET"`
}

type ConfigApp struct {
	LoginRatelimit         int    `envconfig:"LOGIN_RATELIMIT"`
	AccessTokenCookieName  string `envconfig:"ACCESS_TOKEN_COOKIE_NAME"`
	RefreshTokenCookieName string `envconfig:"REFRESH_TOKEN_COOKIE_NAME"`
	AcessTokenExpiration   int    `envconfig:"ACCESS_TOKEN_EXPIRATION"`
	RefreshTokenExpiration int    `envconfig:"REFRESH_TOKEN_EXPIRATION"`
}

type Flags struct {
	Host                                  string
	Port                                  int
	Debug, Verbose, Test                  bool
	AccessLog, ErrorLog                   string
	PrivateKeyPath, PublicKeyPath, JWTKID string
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) ParseFlags(args []string) error {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.StringVar(&c.Flags.Host, "host", "127.0.0.1", "http server host")
	flagSet.IntVar(&c.Flags.Port, "port", 15500, "http server port")
	flagSet.BoolVar(&c.Flags.Debug, "debug", true, "turn on debug captcha, mail")
	flagSet.BoolVar(&c.Flags.Verbose, "verbose", true, "log level = DEBUG")
	flagSet.BoolVar(&c.Flags.Test, "test", false, "for testing (cache in memory)")

	flagSet.StringVar(&c.Flags.AccessLog, "access-log", "", "path to access log file")
	flagSet.StringVar(&c.Flags.ErrorLog, "error-log", "", "path to error log file")

	flagSet.StringVar(
		&c.Flags.PrivateKeyPath,
		"jwt-private-key",
		"",
		"path to private key file",
	)
	flagSet.StringVar(
		&c.Flags.PublicKeyPath,
		"jwt-public-key",
		"",
		"path to public key file",
	)
	flagSet.StringVar(
		&c.Flags.JWTKID,
		"jwt-kid",
		"",
		"path to public key file",
	)

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	return nil
}
