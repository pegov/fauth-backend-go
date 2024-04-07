package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pegov/fauth-backend-go/internal/http/bind"
	"github.com/pegov/fauth-backend-go/internal/http/render"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/service"
)

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request) error
	Get(w http.ResponseWriter, r *http.Request) error
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

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var request *model.LoginRequest
	if err := bind.JSON(r, &request); err != nil {
		return err
	}

	user, err := h.authService.Login(request)
	if err != nil {
		return err
	}

	render.JSON(w, http.StatusOK, user)
	return nil
}

func (h *authHandler) Get(w http.ResponseWriter, r *http.Request) error {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ErrInvalidPathParamType
	}

	user, err := h.authService.Get(id)
	if err != nil {
		return err
	}

	render.JSON(w, http.StatusOK, user)
	return nil
}
