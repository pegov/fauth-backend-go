package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/pegov/fauth-backend-go/internal/api"
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	if err := api.Run(
		ctx,
		os.Args[1:],
		os.Getenv,
		os.Stdout,
		os.Stderr,
		signals,
	); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
