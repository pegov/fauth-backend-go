package api

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pegov/fauth-backend-go/internal/api/handler"
	"github.com/pegov/fauth-backend-go/internal/db"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	db := db.GetDB(os.Getenv("DATABASE_URL"))
	userRepo := repo.NewUserRepo(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)
	authSubRouter := chi.NewRouter()
	authSubRouter.Post("/login", makeHandler(authHandler.Login))
	authSubRouter.Get("/{id}", makeHandler(authHandler.Get))
	r.Mount("/api/v1/users", authSubRouter)

	return r
}
