package dataresponse

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	adapterslog "github.com/raoptimus/data-response.go/pkg/logger/adapter/slog"
	"github.com/raoptimus/data-response.go/v2/formatter"
	"github.com/stretchr/testify/assert"
)

func TestWrite_Success(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))
	resp := DataResponse{
		statusCode: http.StatusOK,
		data:       map[string]string{"key": "value"},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	err := Write(r.Context(), w, resp, factory)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWrite_FormatError(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	// Unserialize data
	resp := DataResponse{
		statusCode: http.StatusOK,
		data:       map[string]any{"bad": make(chan int)},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	err := Write(r.Context(), w, resp, factory)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWrite_ErrorResponseFormatError(t *testing.T) {
	// Создаем formatter, который всегда падает
	badFormatter := &alwaysFailFormatter{}

	factory := New(WithFormatter(badFormatter))
	resp := DataResponse{statusCode: http.StatusOK, data: "test"}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	err := Write(r.Context(), w, resp, factory)

	assert.Error(t, err)

	// Must returns minimal response with error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal Server Error")
}

type alwaysFailFormatter struct {
	BaseFormatter
}

func (f *alwaysFailFormatter) Format(resp DataResponse) (FormattedResponse, error) {
	return FormattedResponse{}, errors.New("always fails")
}

func (f *alwaysFailFormatter) ContentType() string {
	return MimeTypeJSON.String()
}

func TestWrite_MinimalError_WriteFails(t *testing.T) {
	factory := New(
		WithLogger(adapterslog.New(slog.Default())),
		WithFormatter(&alwaysFailFormatter{}),
	)

	// Use ResponseWriter, which failed on call Write
	w := &failingWriter{
		ResponseWriter: httptest.NewRecorder(),
	}
	r := httptest.NewRequest("GET", "/", nil)

	resp := DataResponse{statusCode: http.StatusOK, data: "test"}

	err := Write(r.Context(), w, resp, factory)

	assert.Error(t, err)

	// Проверяем что все попытки были сделаны и залогированы
	// (требует mock logger для проверки)
}

type failingWriter struct {
	http.ResponseWriter
}

func (fw *failingWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write failed")
}

func (fw *failingWriter) WriteHeader(statusCode int) {
	// Simulating a successful WriteHeader
}

func TestWrite_PartialWrite(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	// ResponseWriter, which writes partial data
	w := &partialWriter{
		ResponseWriter: httptest.NewRecorder(),
	}
	r := httptest.NewRequest("GET", "/", nil)

	resp := DataResponse{
		statusCode: http.StatusOK,
		data:       map[string]string{"key": "value"},
	}

	err := Write(r.Context(), w, resp, factory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partial write")
}

type partialWriter struct {
	http.ResponseWriter
}

func (pw *partialWriter) Write(b []byte) (int, error) {
	// Writes the half data
	return len(b) / 2, nil
}
