package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pegov/fauth-backend-go/internal/http/render"
)

type Foo struct {
	A string `json:"a"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var foo Foo
		err := render.BindJSON(r, &foo)
		if err != nil {
			return
		}

		render.JSON(w, http.StatusOK, map[string]any{"message": "ok"})
	})
	http.ListenAndServe("localhost:3000", r)
}
