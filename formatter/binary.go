package formatter

import (
	"path/filepath"

	dr "github.com/raoptimus/data-response.go"
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
	if !resp.IsBinary() {
		return dr.FormattedResponse{}, dr.NewError(500, "response is not binary")
	}

	contentType := resp.ContentType()
	if contentType == "" {
		ext := filepath.Ext(resp.Filename())
		contentType = dr.MimeTypeFromExtension(ext).String()
	}

	// Return stream for writer.go
	return dr.FormattedResponse{
		ContentType: contentType,
		Stream:      resp.Binary(),
		StreamSize:  resp.Size(),
	}, nil
}

// ContentType returns application/octet-stream.
func (f *Binary) ContentType() string {
	return dr.MimeTypeOctetStream.String()
}

// CanFormatBinary returns true.
func (f *Binary) CanFormatBinary() bool {
	return true
}

func init() {
	dr.SetDefaultBinaryFormatter(func() dr.Formatter {
		return NewBinary()
	})
}
