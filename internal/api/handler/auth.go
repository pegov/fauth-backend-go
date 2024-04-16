package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/http/bind"
	"github.com/pegov/fauth-backend-go/internal/http/render"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/service"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request) error
	Login(w http.ResponseWriter, r *http.Request) error
	Token(w http.ResponseWriter, r *http.Request) error
	RefreshToken(w http.ResponseWriter, r *http.Request) error
	Logout(w http.ResponseWriter, r *http.Request) error
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

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var request *model.LoginRequest
	if err := bind.JSON(r, &request); err != nil {
		return err
	}

	accessToken, refreshToken, err := h.authService.Login(request)
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

var (
	ErrNoTokenCookie = errors.New("no token cookie error")
	ErrRenderJSON    = errors.New("render json error")
)

func (h *authHandler) Token(w http.ResponseWriter, r *http.Request) error {
	v, err := r.Cookie("access_c")
	if err != nil {
		return ErrNoTokenCookie
	}

	user, err := h.authService.Token(v.String())
	if err != nil {
		return err
	}

	if err := render.JSON(w, http.StatusOK, user); err != nil {
		return fmt.Errorf("%w: %w", ErrRenderJSON, err)
	}

	return nil
}

func (h *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	v, err := r.Cookie("refresh_c")
	if err != nil {
		return ErrNoTokenCookie
	}

	accessToken, err := h.authService.RefreshToken(v.String())
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

	return nil
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_c",
		Value:    "",
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   -1,
		Secure:   false, // TODO
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_c",
		Value:    "",
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   -1,
		Secure:   false, // TODO
		HttpOnly: true,
	})
	return nil
}
