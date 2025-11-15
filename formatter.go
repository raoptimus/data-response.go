/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"io"
)

// FormattedResponse represents a formatted response ready to be written.
type FormattedResponse struct {
	// ContentType for Content-Type header
	ContentType string

	// Stream formatted content
	Stream io.Reader

	// StreamSize is the size of stream data (-1 if unknown)
	StreamSize int64
}

// Formatter defines the interface for response formatting strategies.
type Formatter interface {
	// Format converts DataResponse to FormattedResponse.
	Format(resp DataResponse) (FormattedResponse, error)

	// ContentType returns the default Content-Type for this formatter.
	ContentType() string

	// CanFormatBinary returns true if formatter can handle binary data.
	CanFormatBinary() bool
}

// BaseFormatter provides common functionality for formatters.
type BaseFormatter struct{}

// CanFormatBinary returns false by default.
func (BaseFormatter) CanFormatBinary() bool {
	return false
}
