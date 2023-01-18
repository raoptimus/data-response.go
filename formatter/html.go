package formatter

import (
	"errors"
	"io"
	"net/http"
)

var ErrDataIsNotStringable = errors.New("data is not a string-able")

type Html struct{}

func NewHtml() *Html {
	return &Html{}
}

func (h *Html) Write(w http.ResponseWriter, statusCode int, data any) error {
	w.WriteHeader(statusCode)

	if s, ok := data.(string); ok {
		_, err := io.WriteString(w, s)
		return err
	}

	if b, ok := data.([]byte); ok {
		_, err := w.Write(b)
		return err
	}

	return ErrDataIsNotStringable
}
