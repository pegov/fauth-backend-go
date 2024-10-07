package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/api/handler"
	"github.com/pegov/fauth-backend-go/internal/http/bind"
	"github.com/pegov/fauth-backend-go/internal/http/render"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/service"
)

type HandlerFuncWithError = func(w http.ResponseWriter, r *http.Request) error

func makeHandler(fn HandlerFuncWithError, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var validationError *model.ValidationError
		var bindJSONError *bind.BindJSONError
		if err := fn(w, r); err != nil {
			switch {
			case errors.Is(err, handler.ErrInvalidPathParamType):
				render.String(w, http.StatusBadRequest, err.Error())

			case errors.Is(err, service.ErrUserNotFound):
				render.String(w, http.StatusNotFound, "Not found")

			case errors.Is(err, service.ErrInvalidCaptcha):
				render.String(w, http.StatusBadRequest, err.Error())

			case errors.Is(err, service.ErrUserAlreadyExistsEmail),
				errors.Is(err, service.ErrUserAlreadyExistsUsername),
				errors.Is(err, service.ErrUserPasswordNotSet),
				errors.As(err, &validationError):
				render.JSON(
					w,
					http.StatusBadRequest,
					map[string]string{"detail": err.Error()},
				)

			case errors.Is(err, service.ErrUserNotActive),
				errors.Is(err, service.ErrPasswordVerification),
				errors.Is(err, handler.ErrNoTokenCookie):
				render.String(w, http.StatusUnauthorized, "Unauthorized")

			case errors.As(err, &bindJSONError):
				// TODO: details
				render.String(
					w,
					http.StatusUnprocessableEntity,
					"Unprocessable entity",
				)

			default:
				logger.Error("Internal server error", slog.Any("err", err))
				render.String(
					w,
					http.StatusInternalServerError,
					"Internal server error",
				)
			}
		}
	}
}
