package config

import (
	"fmt"
	"os"
)

type Config struct {
	Host string
	Port string
}

func New() Config {
	cfg := Config{}

	cfg.Host = readEnv("HOST")
	cfg.Port = readEnv("PORT")

	return cfg
}

func readEnv(name string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		panic(fmt.Sprintf("Failed to load env: %s", name))
	}

	return v
}
