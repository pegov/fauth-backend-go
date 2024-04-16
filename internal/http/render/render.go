package render

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func String(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Write([]byte(message))
}
