package config

import (
	"fmt"
	"os"
)

type Config struct {
	Host            string
	Port            string
	DatabaseURL     string
	CacheURL        string
	RecaptchaSecret string
}

func New() Config {
	cfg := Config{}

	cfg.Host = readEnv("HOST")
	cfg.Port = readEnv("PORT")
	cfg.DatabaseURL = readEnv("DATABASE_URL")
	cfg.CacheURL = readEnv("CACHE_URL")
	cfg.RecaptchaSecret = readEnv("RECAPTCHA_SECRET")

	return cfg
}

func readEnv(name string) string {
	v, ok := os.LookupEnv(name)
	if !ok || v == "" {
		panic(fmt.Sprintf("Failed to load env: %s", name))
	}

	return v
}
