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
	switch v := data.(type) {
	case string:
		if _, err := io.WriteString(w, v); err != nil {
			return err
		}
	case []byte:
		if _, err := w.Write(v); err != nil {
			return err
		}
	default:
		return ErrDataIsNotStringable
	}

	w.WriteHeader(statusCode)

	return nil
}
