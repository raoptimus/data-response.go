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
		err := json.Write(w, http.StatusOK, binary)
		require.NoError(t, err)
		resp := w.Header().Get("Content-Type")
		require.Equal(t, contentType, resp)
	}

	w := httptest.NewRecorder()
	handler(w, nil)
}

func TestBinaryWrite_Error(t *testing.T) {
	data := struct{}{}
	json := NewJson()

	handler := func(w http.ResponseWriter, r *http.Request) {
		err := json.Write(w, http.StatusOK, data)
		require.NoError(t, err)
		resp := w.Header().Get("Content-Type")
		require.NotEqual(t, contentType, resp)
	}

	w := httptest.NewRecorder()
	handler(w, nil)
}
