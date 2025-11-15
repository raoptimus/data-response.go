/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package formatter

import (
	"path/filepath"

	dr "github.com/raoptimus/data-response.go/v2"
)

// Binary is a binary file formatter.
type Binary struct {
	dr.BaseFormatter
}

// NewBinary creates a new binary formatter.
func NewBinary() *Binary {
	return &Binary{}
}

// Format prepares binary data for writing.
func (f *Binary) Format(resp dr.DataResponse) (dr.FormattedResponse, error) {
	contentType := resp.ContentType()
	if contentType == "" {
		ext := filepath.Ext(resp.Filename())
		contentType = dr.MimeTypeFromExtension(ext).String()
	}

	size, body, err := resp.Body()
	if err != nil {
		return dr.FormattedResponse{}, err
	}

	// Return stream for writer.go
	return dr.FormattedResponse{
		ContentType: contentType,
		Stream:     body,
		StreamSize: size,
	}, nil
}

// ContentType returns application/octet-stream.
func (f *Binary) ContentType() string {
	return dr.ContentTypeOctetStream
}

// CanFormatBinary returns true.
func (f *Binary) CanFormatBinary() bool {
	return true
}
