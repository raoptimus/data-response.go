package formatter

import (
	"encoding/xml"
	"net/http"

	dataresponse "github.com/raoptimus/data-response.go"
)

// XML is an XML response formatter.
type XML struct {
	dataresponse.BaseFormatter
	Indent bool
}

// NewXML creates a new XML formatter.
func NewXML() *XML {
	return &XML{Indent: false}
}

// NewXMLIndent creates a new XML formatter with indentation.
func NewXMLIndent() *XML {
	return &XML{Indent: true}
}

// Format writes XML response.
func (f *XML) Format(w http.ResponseWriter, resp dataresponse.DataResponse) error {
	if resp.IsBinary() {
		return dataresponse.NewError(http.StatusInternalServerError, "cannot format binary as XML")
	}

	f.WriteHeaders(w, resp, f.ContentType())
	w.WriteHeader(resp.StatusCode())

	encoder := xml.NewEncoder(w)
	if f.Indent {
		encoder.Indent("", "  ")
	}

	return encoder.Encode(resp)
}

// ContentType returns application/xml.
func (f *XML) ContentType() string {
	return dataresponse.MimeTypeXML.String()
}
