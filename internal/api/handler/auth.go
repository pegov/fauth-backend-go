package handler

import (
	"errors"
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/config"
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
	Me(w http.ResponseWriter, r *http.Request) error
}

type authHandler struct {
	cfg         *config.Config
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

func (h *authHandler) setAccessCookie(
	w http.ResponseWriter,
	accessToken string,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.AccessTokenCookieName,
		Value:    accessToken,
		Path:     "/",
		Domain:   h.cfg.HttpDomain,
		MaxAge:   h.cfg.AcessTokenExpiration,
		Secure:   h.cfg.HttpSecure,
		HttpOnly: true,
	})
}

func (h *authHandler) setCookies(
	w http.ResponseWriter,
	accessToken,
	refreshToken string,
) {
	h.setAccessCookie(w, accessToken)
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.RefreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		Domain:   h.cfg.HttpDomain,
		MaxAge:   h.cfg.RefreshTokenExpiration,
		Secure:   h.cfg.HttpSecure,
		HttpOnly: true,
	})
}

func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var request *model.RegisterRequest
	if err := bind.JSON(r, &request); err != nil {
		return err
	}

	accessToken, refreshToken, err := h.authService.Register(r.Context(), request)
	if err != nil {
		return err
	}

	h.setCookies(w, accessToken, refreshToken)
	return nil
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var request *model.LoginRequest
	if err := bind.JSON(r, &request); err != nil {
		return err
	}

	accessToken, refreshToken, err := h.authService.Login(r.Context(), request)
	if err != nil {
		return err
	}

	h.setCookies(w, accessToken, refreshToken)
	return nil
}

var (
	ErrNoTokenCookie = errors.New("no token cookie error")
)

func (h *authHandler) Token(w http.ResponseWriter, r *http.Request) error {
	v, err := r.Cookie(h.cfg.AccessTokenCookieName)
	if err != nil {
		return ErrNoTokenCookie
	}

	user, err := h.authService.Token(r.Context(), v.String())
	if err != nil {
		return err
	}

	return render.JSON(w, http.StatusOK, user)
}

func (h *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	v, err := r.Cookie("refresh_c")
	if err != nil {
		return ErrNoTokenCookie
	}

	accessToken, err := h.authService.RefreshToken(r.Context(), v.String())
	if err != nil {
		return err
	}

	h.setAccessCookie(w, accessToken)
	return nil
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.AccessTokenCookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cfg.HttpDomain,
		MaxAge:   -1,
		Secure:   h.cfg.HttpSecure,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cfg.HttpDomain,
		MaxAge:   -1,
		Secure:   h.cfg.HttpSecure,
		HttpOnly: true,
	})
	return nil
}

func (h *authHandler) Me(w http.ResponseWriter, r *http.Request) error {
	v, err := r.Cookie(h.cfg.AccessTokenCookieName)
	if err != nil {
		return ErrNoTokenCookie
	}

	tokenPayload, err := h.authService.Token(r.Context(), v.String())
	if err != nil {
		return err
	}

	me, err := h.authService.Me(r.Context(), tokenPayload.ID)
	if err != nil {
		return err
	}

	return render.JSON(w, http.StatusOK, me)
}
