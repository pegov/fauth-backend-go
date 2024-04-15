package api

import (
	"os"

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
	r.Use(middleware.Logger)

	db := db.GetDB(os.Getenv("DATABASE_URL"))
	userRepo := repo.NewUserRepo(db)
	captchaClient := captcha.NewReCaptchaClient(os.Getenv("RECAPTCHA_SECRET"))
	passwordHasher := password.NewBcryptPasswordHasher()
	tokenBackend := token.NewJwtBackend([]byte("TODO"), []byte("TODO"), "TODO_KID")
	authService := service.NewAuthService(userRepo, captchaClient, passwordHasher, tokenBackend)
	authHandler := handler.NewAuthHandler(authService)
	authSubRouter := chi.NewRouter()
	authSubRouter.Post("/register", makeHandler(authHandler.Register))
	authSubRouter.Post("/login", makeHandler(authHandler.Login))
	r.Mount("/api/v1/users", authSubRouter)

	return r
}
