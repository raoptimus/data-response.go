package response

import (
	"net/http"
)

type (
	DataResponse struct {
		data       any
		statusCode int
		header     http.Header
	}
)

func NewDataResponse(code int, data any) *DataResponse {
	return &DataResponse{
		data:       data,
		statusCode: code,
		header:     http.Header{},
	}
}

func (d *DataResponse) Header() http.Header {
	return d.header
}

func (d *DataResponse) Data() any {
	return d.data
}

func (d *DataResponse) StatusCode() int {
	return d.statusCode
}

func (d *DataResponse) HeaderValues(key string) []string {
	return d.header.Values(key)
}

func (d *DataResponse) HeaderLine(key string) string {
	return d.header.Get(key)
}

func (d *DataResponse) WithHeader(key, value string) *DataResponse {
	d.header.Add(key, value)
	return d
}

func (d *DataResponse) HasHeader(key string) bool {
	return d.header.Get(key) != ""
}
