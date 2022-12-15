package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	response "github.com/raoptimus/data-response.go"
	"github.com/stretchr/testify/assert"

	"github.com/raoptimus/data-response.go/formatter"
)

type testHandler struct{}

func (s *testHandler) Handle(f response.Factory, r *http.Request) *response.DataResponse {
	return f.CreateResponse(r.Method, 200)
}

func TestHandle_GetStdRequest_ReturnsResponseSuccessfully(t *testing.T) {
	f := response.NewFactory(formatter.NewJsonPretty(), false)
	h := DataResponse(&testHandler{}, f)
	srv := httptest.NewServer(h)
	defer srv.Close()

	c := &http.Client{Transport: &http.Transport{}}
	resp, err := c.Get("http://" + srv.Listener.Addr().String())
	assert.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var content string
	err = json.NewDecoder(resp.Body).Decode(&content)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, content)
}
