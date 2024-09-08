package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func getenv(s string) string {
	switch s {
	case "DATABASE_URL",
		"CACHE_URL",
		"RECAPTCHA_SECRET",
		"HTTP_DOMAIN",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"SMTP_HOST",
		"SMTP_PORT":
		return "-"
	case "DATABASE_MAX_IDLE_CONNS",
		"DATABASE_MAX_OPEN_CONNS",
		"DATABASE_CONN_MAX_LIFETIME",
		"LOGIN_RATELIMIT",
		"ACCESS_TOKEN_EXPIRATION",
		"REFRESH_TOKEN_EXPIRATION":
		return "10"
	case "HTTP_SECURE":
		return "0"
	case "ACCESS_TOKEN_COOKIE_NAME":
		return "access"
	case "REFRESH_TOKEN_COOKIE_NAME":
		return "refresh"
	}

	return ""
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	args := []string{
		"--test",
	}
	handler, _, _, _, err := Prepare(
		ctx,
		args,
		getenv,
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
