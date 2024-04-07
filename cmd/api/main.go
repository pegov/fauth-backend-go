package main

import (
	"log"
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/api"
)

func main() {
	r := api.NewRouter()
	log.Println("Starting server...")
	http.ListenAndServe("localhost:3000", r)
}
