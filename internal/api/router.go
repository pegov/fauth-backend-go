package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pegov/fauth-backend-go/internal/api/handler"
	"github.com/pegov/fauth-backend-go/internal/service"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	authService := service.NewAuthService()
	authHandler := handler.NewAuthHandler(authService)
	authSubRouter := chi.NewRouter()
	authSubRouter.Post("/login", makeHandler(authHandler.Login))
	r.Mount("/api/v1/users", authSubRouter)

	return r
}
