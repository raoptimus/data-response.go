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

	json "github.com/json-iterator/go"
	"github.com/raoptimus/data-response.go/v2/response"
)

// JSON is a JSON response formatter.
type JSON struct {
	response.BaseFormatter
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

// Format converts DataResponse to formatted JSON.
func (f *JSON) Format(resp *response.DataResponse) (response.FormattedResponse, error) {
	if resp.IsBinary() {
		return response.FormattedResponse{}, response.NewError(errCode500, "cannot format binary as JSON")
	}

	data := resp.Data()

	if data == nil {
		body := []byte("null")

		return response.FormattedResponse{
			Stream:     bytes.NewReader(body),
			StreamSize: int64(len(body)),
		}, nil
	}

	// Serialize to buffer
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(data); err != nil {
		return response.FormattedResponse{}, response.WrapError(errCode500, err, "failed to encode JSON")
	}

	return response.FormattedResponse{
		Stream:     bytes.NewReader(buf.Bytes()),
		StreamSize: int64(buf.Len()),
	}, nil
}

// ContentType returns application/json.
func (f *JSON) ContentType() string {
	return response.ContentTypeJSON
}
