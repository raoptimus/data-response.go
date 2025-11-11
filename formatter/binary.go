package formatter

import (
	"io"
	"net/http"
	"path/filepath"

	dataresponse "github.com/raoptimus/data-response.go"
)

// Binary is a binary file formatter with efficient streaming.
type Binary struct {
	dataresponse.BaseFormatter
	BufferSize int
}

// NewBinary creates a new binary formatter with default 32KB buffer.
func NewBinary() *Binary {
	return &Binary{
		BufferSize: 32 * 1024,
	}
}

// Format writes binary response with efficient buffered copying.
func (f *Binary) Format(w http.ResponseWriter, resp dataresponse.DataResponse) error {
	if !resp.IsBinary() {
		return dataresponse.NewError(http.StatusInternalServerError, "response is not binary")
	}

	contentType := resp.ContentType()
	if contentType == "" {
		// Use MimeTypeFromExtension
		ext := filepath.Ext(resp.Filename())
		contentType = dataresponse.MimeTypeFromExtension(ext).String()
	}

	f.WriteHeaders(w, resp, contentType)
	w.WriteHeader(resp.StatusCode())

	if resp.Size() > 0 {
		_, err := io.CopyN(w, resp.Binary(), resp.Size())
		return err
	}

	buf := make([]byte, f.BufferSize)
	_, err := io.CopyBuffer(w, resp.Binary(), buf)
	return err
}

// ContentType returns application/octet-stream.
func (f *Binary) ContentType() string {
	return dataresponse.MimeTypeOctetStream.String()
}

// CanFormatBinary returns true.
func (f *Binary) CanFormatBinary() bool {
	return true
}

func init() {
	dataresponse.SetDefaultBinaryFormatter(func() dataresponse.Formatter {
		return NewBinary()
	})
}
