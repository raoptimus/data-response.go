package formatter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonWrite_Success(t *testing.T) {
	contentType := "application/some"
	binary := NewBinaryData([]byte("success"), "some_name", contentType)
	json := NewJson()

	handler := func(w http.ResponseWriter, r *http.Request) {
		err := json.Write(w, http.StatusOK, binary)
		require.NoError(t, err)
		resp := w.Header().Get("Content-Type")
		require.Equal(t, contentType, resp)
	}

	w := httptest.NewRecorder()
	handler(w, nil)
}

func TestJsonWrite_Error(t *testing.T) {
	contentType := "application/some"
	binary := *NewBinaryData([]byte("success"), "some_name", contentType)
	json := NewJson()

	handler := func(w http.ResponseWriter, r *http.Request) {
		err := json.Write(w, http.StatusOK, binary)
		require.NoError(t, err)
		resp := w.Header().Get("Content-Type")
		require.NotEqual(t, contentType, resp)
	}

	w := httptest.NewRecorder()
	handler(w, nil)
}
