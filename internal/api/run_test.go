package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/model"
)

var (
	cfg *config.Config
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	cfg = &config.Config{
		Database: config.ConfigDatabase{
			URL:             os.Getenv("DATABASE_URL"),
			MaxIdleConns:    10,
			MaxOpenConns:    10,
			ConnMaxLifetime: 10,
		},
		Cache: config.ConfigCache{
			URL: "-",
		},
		Captcha: config.ConfigCaptcha{
			RecaptchaSecret: "-",
		},
		HTTP: config.ConfigHTTP{
			Domain: "-",
			Secure: false,
		},
		SMTP: config.ConfigSMTP{
			Username: "-",
			Password: "-",
			Host:     "-",
			Port:     "-",
		},
		App: config.ConfigApp{
			LoginRatelimit:         10,
			AccessTokenCookieName:  "access",
			RefreshTokenCookieName: "refresh",
			AcessTokenExpiration:   10,
			RefreshTokenExpiration: 10,
		},
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL_POSTGRES"))
	if err != nil {
		fmt.Printf("failed to open postgres connection")
		os.Exit(1)
	}
	if _, err := db.Exec(ctx, "DROP DATABASE IF EXISTS fauth_test"); err != nil {
		fmt.Printf("failed to drop database: %s", err)
		os.Exit(1)
	}
	if _, err := db.Exec(ctx, "CREATE DATABASE fauth_test"); err != nil {
		fmt.Printf("failed to create database: %s", err)
		os.Exit(1)
	}
	if err := db.Close(ctx); err != nil {
		fmt.Printf("failed to close postgres connection")
		os.Exit(1)
	}

	db, err = pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("failed to open fauth connection")
		os.Exit(1)
	}
	if _, err := db.Exec(ctx, "DROP TABLE IF EXISTS auth_user"); err != nil {
		fmt.Printf("failed to drop user")
		os.Exit(1)
	}
	if _, err := db.Exec(ctx, "DROP TABLE IF EXISTS auth_oauth"); err != nil {
		fmt.Printf("failed to drop oauth")
		os.Exit(1)
	}
	if err := db.Close(ctx); err != nil {
		fmt.Printf("failed to close fauth connection")
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	args := []string{
		"--test",
	}
	if err := cfg.ParseFlags(args); err != nil {
		t.Fatalf("failed to parse flags: %s", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler, err := PrepareForTest(
		ctx,
		cfg,
		logger,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		t.Fatalf("failed to prepare handler: %s", err)
	}

	server := httptest.NewServer(handler)
	res, err := http.Get(server.URL + "/ping")
	if err != nil {
		t.Fatalf("failed to get response: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read body: %s", err)
	}
	defer res.Body.Close()

	expected := "pong"
	if string(b) != expected {
		t.Fatalf("expected %s, got %s", expected, b)
	}
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	args := []string{
		"--test",
	}
	if err := cfg.ParseFlags(args); err != nil {
		t.Fatalf("failed to parse flags: %s", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler, err := PrepareForTest(
		ctx,
		cfg,
		logger,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		t.Fatalf("failed to prepare handler: %s", err)
	}

	server := httptest.NewServer(handler)
	data := model.RegisterRequest{
		Email:     "test@test.com",
		Username:  "test",
		Password1: "test1234",
		Password2: "test1234",
	}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatalf("failed to serialize json: %s", err)
	}
	req, err := http.NewRequest("POST", server.URL+"/api/v1/users/register", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("failed to build request: %s", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get response: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}
