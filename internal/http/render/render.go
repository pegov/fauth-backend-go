package render

import (
	"encoding/json"
	"net/http"
)

func JSON[T any](w http.ResponseWriter, status int, payload T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func String(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func Status(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}
