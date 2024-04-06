package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

type RenderBindJSONError struct {
	Status  int
	Message string
}

func (e *RenderBindJSONError) Error() string {
	return e.Message
}

func BindJSON(r *http.Request, model any) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			return &RenderBindJSONError{
				http.StatusUnsupportedMediaType,
				"Content-Type header is not \"application/json\"",
			}
		}
	}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(model)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return &RenderBindJSONError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset),
			}

		case errors.Is(err, io.ErrUnexpectedEOF):
			return &RenderBindJSONError{
				Status:  http.StatusBadRequest,
				Message: "Request body contains badly-formed JSON",
			}

		case errors.As(err, &unmarshalTypeError):
			return &RenderBindJSONError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset),
			}

		case errors.Is(err, io.EOF):
			return &RenderBindJSONError{
				Status:  http.StatusBadRequest,
				Message: "Request body must not be empty",
			}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return &RenderBindJSONError{
			Status:  http.StatusBadRequest,
			Message: "Request body must only contain a single JSON object",
		}
	}

	return nil
}
