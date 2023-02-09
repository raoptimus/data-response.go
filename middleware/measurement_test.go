package middleware

import (
	"net/http"
	"net/url"
	"testing"

	httpt "github.com/stretchr/testify/http"
	"github.com/stretchr/testify/mock"
)

func TestMeasurement_ReceiveValidMetrics(t *testing.T) {
	m := NewMockMetricsService(t)
	m.On("Responded", mock.MatchedBy(func(i interface{}) bool {
		data := i.(MetricsData)
		return data.StatusCode == http.StatusInternalServerError &&
			data.Route == "/test" &&
			data.Method == http.MethodGet
	}))
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})
	mh := Measurement(h, m, func(r *http.Request) string {
		return "/test"
	})
	mh.ServeHTTP(&httpt.TestResponseWriter{}, &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{},
	})
}
