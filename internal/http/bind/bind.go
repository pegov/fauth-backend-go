package bind

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type BindJSONError struct {
	Status  int
	Message string
}

func (e *BindJSONError) Error() string {
	return e.Message
}

func JSON[T any](r *http.Request, model T) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			return &BindJSONError{
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
			return &BindJSONError{
				Status: http.StatusBadRequest,
				Message: fmt.Sprintf(
					"Request body contains badly-formed JSON (at position %d)",
					syntaxError.Offset,
				),
			}

		case errors.Is(err, io.ErrUnexpectedEOF):
			return &BindJSONError{
				Status:  http.StatusBadRequest,
				Message: "Request body contains badly-formed JSON",
			}

		case errors.As(err, &unmarshalTypeError):
			return &BindJSONError{
				Status: http.StatusBadRequest,
				Message: fmt.Sprintf(
					"Request body contains an invalid value for the %q field (at position %d)",
					unmarshalTypeError.Field,
					unmarshalTypeError.Offset,
				),
			}

		case errors.Is(err, io.EOF):
			return &BindJSONError{
				Status:  http.StatusBadRequest,
				Message: "Request body must not be empty",
			}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return &BindJSONError{
			Status:  http.StatusBadRequest,
			Message: "Request body must only contain a single JSON object",
		}
	}

	return nil
}
