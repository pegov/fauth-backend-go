package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/pegov/fauth-backend-go/internal/api/handler"
	"github.com/pegov/fauth-backend-go/internal/config"
	"github.com/pegov/fauth-backend-go/internal/log"
	"github.com/pegov/fauth-backend-go/internal/service"
)

func NewRouter(
	cfg *config.Config,
	logger log.Logger,
	authService service.AuthService,
	adminService service.AdminService,
) (chi.Router, error) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(20 * time.Second))

	apiV1Router := chi.NewRouter()

	localMakeHandler := func(handler HandlerFuncWithError) http.HandlerFunc {
		return makeHandler(handler, logger)
	}

	apiV1Router.Group(func(r chi.Router) {
		authHandler := handler.NewAuthHandler(authService)
		r.Post("/register", localMakeHandler(authHandler.Register))
		r.Post("/login", localMakeHandler(authHandler.Login))
		r.Post("/logout", localMakeHandler(authHandler.Logout))
		r.Post("/token", localMakeHandler(authHandler.Token))
		r.Post("/token/refresh", localMakeHandler(authHandler.RefreshToken))
		r.Post("/me", localMakeHandler(authHandler.Me))
	})

	apiV1Router.Group(func(r chi.Router) {
		adminHandler := handler.NewAdminHandler(adminService)
		r.Get("/mass_logout", localMakeHandler(adminHandler.GetMassLogout))
		r.Post("/mass_logout", localMakeHandler(adminHandler.ActivateMassLogout))
		r.Delete("/mass_logout", localMakeHandler(adminHandler.DeactivateMassLogout))
		r.Post("/{id}/ban", localMakeHandler(adminHandler.Ban))
		r.Post("/{id}/unban", localMakeHandler(adminHandler.Unban))
		r.Post("/{id}/kick", localMakeHandler(adminHandler.Kick))
		r.Post("/{id}/unkick", localMakeHandler(adminHandler.Unkick))
	})

	r.Mount("/api/v1/users", apiV1Router)

	return r, nil
}
