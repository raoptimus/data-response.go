package formatter

import (
	"net/http"

	json "github.com/json-iterator/go"
	dataresponse "github.com/raoptimus/data-response.go"
)

// JSON is a JSON response formatter.
type JSON struct {
	dataresponse.BaseFormatter
	Indent bool
}

// NewJSON creates a new JSON formatter.
func NewJSON() *JSON {
	return &JSON{Indent: false}
}

// NewJSONIndent creates a new JSON formatter with pretty-printing.
func NewJSONIndent() *JSON {
	return &JSON{Indent: true}
}

// Format writes JSON response.
func (f *JSON) Format(w http.ResponseWriter, resp dataresponse.DataResponse) error {
	if resp.IsBinary() {
		return dataresponse.NewError(http.StatusInternalServerError, "cannot format binary as JSON")
	}

	f.WriteHeaders(w, resp, f.ContentType())
	w.WriteHeader(resp.StatusCode())

	// Create response structure
	output := map[string]any{
		"status": resp.StatusCode(),
	}

	if resp.Data() != nil {
		output["data"] = resp.Data()
	}

	encoder := json.NewEncoder(w)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(output)
}

// ContentType returns application/json.
func (f *JSON) ContentType() string {
	return dataresponse.MimeTypeJSON.String()
}
