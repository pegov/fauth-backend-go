package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL             string `env:"DATABASE_URL"`
	DatabaseMaxIdleConns    int    `env:"DATABASE_MAX_IDLE_CONNS"`
	DatabaseMaxOpenConns    int    `env:"DATABASE_MAX_OPEN_CONNS"`
	DatabaseConnMaxLifetime int    `env:"DATABASE_CONN_MAX_LIFETIME"`
	CacheURL                string `env:"CACHE_URL"`
	RecaptchaSecret         string `env:"RECAPTCHA_SECRET"`
	HttpDomain              string `env:"HTTP_DOMAIN"`
	HttpSecure              bool   `env:"HTTP_SECURE"`
	LoginRatelimit          int    `env:"LOGIN_RATELIMIT"`
	AccessTokenCookieName   string `env:"ACCESS_TOKEN_COOKIE_NAME"`
	RefreshTokenCookieName  string `env:"REFRESH_TOKEN_COOKIE_NAME"`
	AcessTokenExpiration    int    `env:"ACCESS_TOKEN_EXPIRATION"`
	RefreshTokenExpiration  int    `env:"REFRESH_TOKEN_EXPIRATION"`
	SMTPUsername            string `env:"SMTP_USERNAME"`
	SMTPPassword            string `env:"SMTP_PASSWORD"`
	SMTPHost                string `env:"SMTP_HOST"`
	SMTPPort                string `env:"SMTP_PORT"`
}

func New(getenv func(string) string) (*Config, error) {
	cfg := Config{}
	var missingEnvs []string
	var wrongEnvs []string

	v := reflect.ValueOf(cfg)
	fields := reflect.VisibleFields(v.Type())
	for _, field := range fields {
		f, _ := reflect.TypeOf(cfg).FieldByName(field.Name)
		name := f.Tag.Get("env")
		value := reflect.ValueOf(cfg).FieldByName(f.Name).Interface()
		switch value.(type) {
		case string:
			v, ok := readString(getenv, name)
			if ok {
				reflect.ValueOf(&cfg).Elem().FieldByName(f.Name).Set(reflect.ValueOf(v))
			} else {
				missingEnvs = append(missingEnvs, name)
			}
		case bool:
			v, ok := readString(getenv, name)
			if ok {
				v := v == "1" || v == "True" || v == "TRUE" || v == "true" || v == "yes"
				reflect.ValueOf(&cfg).Elem().FieldByName(f.Name).Set(reflect.ValueOf(v))
			} else {
				missingEnvs = append(missingEnvs, name)
			}
		case int:
			v, ok := readString(getenv, name)
			if ok {
				if n, err := strconv.Atoi(v); err == nil {
					reflect.ValueOf(&cfg).Elem().FieldByName(f.Name).Set(reflect.ValueOf(n))
				} else {
					wrongEnvs = append(wrongEnvs, name)
				}
			} else {
				missingEnvs = append(missingEnvs, name)
			}
		}
	}

	if len(missingEnvs) > 0 || len(wrongEnvs) > 0 {
		var s string
		if len(missingEnvs) > 0 {
			s += fmt.Sprintf("missing envs: %s", strings.Join(missingEnvs, ", "))
		}
		if len(wrongEnvs) > 0 {
			if len(missingEnvs) > 0 {
				s += ", "
			}
			s += fmt.Sprintf("wrong envs: %s", strings.Join(missingEnvs, ", "))
		}
		return &cfg, fmt.Errorf(s)
	}

	return &cfg, nil
}

func (cfg *Config) Pretty() []byte {
	prettyCfg, _ := json.MarshalIndent(cfg, "", "\t")
	return prettyCfg
}

func readString(getenv func(string) string, name string) (string, bool) {
	v := getenv(name)
	if v == "" {
		return "", false
	}

	return v, true
}
