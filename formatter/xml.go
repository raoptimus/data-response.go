/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package formatter

import (
	"bytes"
	"encoding/xml"

	"github.com/raoptimus/data-response.go/v2/response"
)

// XML is an XML response formatter.
type XML struct {
	response.BaseFormatter
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
func (f *XML) Format(resp *response.DataResponse) (response.FormattedResponse, error) {
	if resp.IsBinary() {
		return response.FormattedResponse{}, response.NewError(errCode500, "cannot format binary as XML")
	}

	// Serialize only resp.Data()
	data := resp.Data()

	if data == nil {
		body := []byte("")
		return response.FormattedResponse{
			Stream:     bytes.NewReader(body),
			StreamSize: int64(len(body)),
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
		return response.FormattedResponse{}, response.WrapError(errCode500, err, "failed to encode XML")
	}

	if err := encoder.Flush(); err != nil {
		return response.FormattedResponse{}, response.WrapError(errCode500, err, "failed to flush XML encoder")
	}

	return response.FormattedResponse{
		Stream:     bytes.NewReader(buf.Bytes()),
		StreamSize: int64(buf.Len()),
	}, nil
}

// ContentType returns application/xml.
func (f *XML) ContentType() string {
	return response.ContentTypeXML
}
