package config

import "time"

type Config struct {
	Database Database `flag:"db" env:"DB"`
	Cache    Cache
	HTTP     HTTP
	SMTP     SMTP
	Captcha  Captcha
	OAuth    OAuth `flag:"oauth" env:"OAUTH"`
	App      App
	Flags    Flags `flag:"" env:""`
}

type Database struct {
	URL             string        `usage:"url для postgres"`
	MaxIdleConns    int           `default:"20"`
	MaxOpenConns    int           `default:"20"`
	ConnMaxLifetime time.Duration `default:"1m"`
}

type Cache struct {
	URL string
}

type HTTP struct {
	Domain string
	Secure bool
}

type SMTP struct {
	Username string
	Password string
	Host     string
	Port     string
}

type Captcha struct {
	RecaptchaSecret string `cli:"optional"`
}

type OAuth struct {
	Providers          []string `cli:"optional"`
	GoogleClientID     string   `cli:"optional"`
	GoogleClientSecret string   `cli:"optional"`
	VKAppID            string   `cli:"optional"`
	VKAppSecret        string   `cli:"optional"`
}

type App struct {
	LoginRatelimit         int
	AccessTokenCookieName  string `default:"access"`
	RefreshTokenCookieName string `default:"refresh"`
	AccessTokenExpiration  int
	RefreshTokenExpiration int
}

type Flags struct {
	Host                                  string `usage:"host for api server"`
	Port                                  int    `default:"3000"`
	Debug                                 bool   `env:"-"`
	Verbose                               bool   `env:"-"`
	Test                                  bool   `env:"-"`
	AccessLog, ErrorLog                   string `cli:"optional"`
	PrivateKeyPath, PublicKeyPath, JWTKID string
}
