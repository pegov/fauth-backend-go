package api

import (
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/pegov/fauth-backend-go/internal/api/handler"
	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/db"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
	"github.com/pegov/fauth-backend-go/internal/token"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Second))

	dbClient := db.GetDB(os.Getenv("DATABASE_URL"))
	cacheClient := db.GetCache(os.Getenv("CACHE_URL"))
	userRepo := repo.NewUserRepo(dbClient, cacheClient)
	captchaClient := captcha.NewReCaptchaClient(os.Getenv("RECAPTCHA_SECRET"))
	passwordHasher := password.NewBcryptPasswordHasher()
	privateKey, err := os.ReadFile("./id_ed25519_auth_1.key")
	if err != nil {
		panic(err)
	}
	publicKey, err := os.ReadFile("./id_ed25519_auth_1.pub")
	if err != nil {
		panic(err)
	}
	tokenBackend := token.NewJwtBackend(privateKey, publicKey, "1")

	apiV1Router := chi.NewRouter()

	apiV1Router.Group(func(r chi.Router) {
		authService := service.NewAuthService(userRepo, captchaClient, passwordHasher, tokenBackend)
		authHandler := handler.NewAuthHandler(authService)
		r.Post("/register", makeHandler(authHandler.Register))
		r.Post("/login", makeHandler(authHandler.Login))
		r.Post("/token", makeHandler(authHandler.Login))
		r.Post("/token/refresh", makeHandler(authHandler.Login))
		r.Post("/me", makeHandler(authHandler.Me))
	})

	apiV1Router.Group(func(r chi.Router) {
		adminService := service.NewAdminService(userRepo)
		adminHandler := handler.NewAdminHandler(adminService)
		r.Post("/{id}/ban", makeHandler(adminHandler.Ban))
		r.Post("/{id}/unban", makeHandler(adminHandler.Unban))
		r.Post("/{id}/kick", makeHandler(adminHandler.Kick))
		r.Post("/{id}/unkick", makeHandler(adminHandler.Unkick))
	})

	r.Mount("/api/v1/users", apiV1Router)

	return r
}
