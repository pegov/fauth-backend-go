package api

import (
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/http/render"
)

type HandlerFuncWithError = func(w http.ResponseWriter, r *http.Request) error

func makeHandler(fn HandlerFuncWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			switch err {
			default:
				render.String(w, http.StatusInternalServerError, "Internal server error")
			}
		}
	}
}
