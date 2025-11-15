/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package formatter

import (
	"bufio"
	"bytes"

	json "github.com/json-iterator/go"
	dr "github.com/raoptimus/data-response.go/v2"
)

// JSON is a JSON response formatter.
type JSON struct {
	dr.BaseFormatter
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
func (f *JSON) Format(resp dr.DataResponse) (dr.FormattedResponse, error) {
	if resp.IsBinary() {
		return dr.FormattedResponse{}, dr.NewError(500, "cannot format binary as JSON")
	}

	data := resp.Data()

	if data == nil {
		var buf bytes.Buffer
		buf.Write([]byte("null"))

		return dr.FormattedResponse{
			ContentType: f.ContentType(),
			Stream:     bufio.NewReader(&buf),
			StreamSize: int64(buf.Len()),
		}, nil
	}

	// Serialize to buffer
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(data); err != nil {
		return dr.FormattedResponse{}, dr.WrapError(500, err, "failed to encode JSON")
	}

	return dr.FormattedResponse{
		ContentType: f.ContentType(),
		Stream:     bufio.NewReader(&buf),
		StreamSize: int64(buf.Len()),
	}, nil
}

// ContentType returns application/json.
func (f *JSON) ContentType() string {
	return dr.ContentTypeJSON
}
