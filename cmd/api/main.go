package main

import (
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/api"
)

func main() {
	r := api.NewRouter()
	http.ListenAndServe("localhost:3000", r)
}
