package config

import (
	"os"
)

type Config struct {
	Host string
	Port string
}

func New() Config {
	cfg := Config{}

	host, ok := os.LookupEnv("HOST")
	if !ok {
		panic("Failed to load env: HOST")
	}

	cfg.Host = host

	port, ok := os.LookupEnv("PORT")
	if !ok {
		panic("Failed to load env: HOST")
	}

	cfg.Port = port

	return cfg
}
