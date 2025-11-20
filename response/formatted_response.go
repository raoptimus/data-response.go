package response

import "io"

// FormattedResponse represents a formatted response ready to be written.
type FormattedResponse struct {
	// Stream formatted content
	Stream io.Reader

	// StreamSize is the size of stream data (-1 if unknown)
	StreamSize int64
}
