package handler

import (
	"errors"
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/http/bind"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/service"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request) error
}

type authHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *authHandler {
	return &authHandler{
		authService: authService,
	}
}

var (
	ErrInvalidPathParamType = errors.New("invalid path param type")
)

func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var request *model.RegisterRequest
	if err := bind.JSON(r, &request); err != nil {
		return err
	}

	accessToken, refreshToken, err := h.authService.Register(request)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_c",
		Value:    accessToken,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   60 * 60 * 6,
		Secure:   false, // TODO
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_c",
		Value:    refreshToken,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   60 * 60 * 24 * 31,
		Secure:   false, // TODO
		HttpOnly: true,
	})

	return nil
}
