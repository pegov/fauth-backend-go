package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/api/handler"
	"github.com/pegov/fauth-backend-go/internal/http/render"
	"github.com/pegov/fauth-backend-go/internal/service"
)

type HandlerFuncWithError = func(w http.ResponseWriter, r *http.Request) error

func makeHandler(fn HandlerFuncWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			switch {
			case errors.Is(err, handler.ErrInvalidPathParamType):
				render.String(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrUserNotFound):
				render.String(w, http.StatusNotFound, "Not found")
			default:
				log.Println(err)
				render.String(w, http.StatusInternalServerError, "Internal server error")
			}
		}
	}
}
