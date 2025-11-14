package formatter

import (
	"bytes"
	"encoding/xml"

	dr "github.com/raoptimus/data-response.go"
)

// XML is an XML response formatter.
type XML struct {
	dr.BaseFormatter
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

// Format converts DataResponse to formatted XML.
func (f *XML) Format(resp dr.DataResponse) (dr.FormattedResponse, error) {
	if resp.IsBinary() {
		return dr.FormattedResponse{}, dr.NewError(500, "cannot format binary as XML")
	}

	// Serialize only resp.Data()
	data := resp.Data()

	if data == nil {
		return dr.FormattedResponse{
			ContentType: f.ContentType(),
			Body:        []byte(""),
		}, nil
	}

	var buf bytes.Buffer

	// Add XML header
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	if f.Indent {
		encoder.Indent("", "  ")
	}

	if err := encoder.Encode(data); err != nil {
		return dr.FormattedResponse{}, dr.WrapError(500, err, "failed to encode XML")
	}

	if err := encoder.Flush(); err != nil {
		return dr.FormattedResponse{}, dr.WrapError(500, err, "failed to flush XML encoder")
	}

	return dr.FormattedResponse{
		ContentType: f.ContentType(),
		Body:        buf.Bytes(),
	}, nil
}

// ContentType returns application/xml.
func (f *XML) ContentType() string {
	return dr.MimeTypeXML.String()
}
