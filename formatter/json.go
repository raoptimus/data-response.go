package formatter

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	pretty bool
}

func NewJson() *Json {
	return &Json{pretty: false}
}
func NewJsonPretty() *Json {
	return &Json{pretty: true}
}

func (j *Json) Write(data any, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if j.pretty {
		enc.SetIndent("", "    ")
	}

	return enc.Encode(data)
}

func (j *Json) Pretty() *Json {
	j.pretty = true
	return j
}
