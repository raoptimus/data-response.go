package formatter

import (
	"net/http"

	"github.com/pkg/errors"
)

var ErrDataIsNotStringable = errors.New("data is not a string-able")

type Html struct{}

func NewHtml() *Html {
	return &Html{}
}

func (h *Html) Marshal(_ http.Header, data any) ([]byte, error) {
	switch v := data.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, errors.WithStack(ErrDataIsNotStringable)
	}
}
