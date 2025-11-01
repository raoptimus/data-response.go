package formatter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	contentType = "application/some"
)

func TestBinaryWrite_Success(t *testing.T) {
	binary := NewBinaryData([]byte("success"), "some_name", contentType)
	json := NewJson()

	handler := func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		bytes, err := json.Marshal(header, binary)
		require.NoError(t, err)
		require.NotEmpty(t, bytes)

		resp := header.Get("Content-Type")
		require.Equal(t, contentType, resp)
	}

	w := httptest.NewRecorder()
	handler(w, nil)
}

func TestBinaryWrite_Error(t *testing.T) {
	data := struct{}{}
	json := NewJson()

	handler := func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		bytes, err := json.Marshal(header, data)
		require.NoError(t, err)
		require.NotEmpty(t, bytes)

		resp := header.Get("Content-Type")
		require.NotEqual(t, contentType, resp)
	}

	w := httptest.NewRecorder()
	handler(w, nil)
}
