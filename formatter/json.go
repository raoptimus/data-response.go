package formatter

import (
	"encoding/json"
	"html"
	"net/http"
)

type Json struct {
	pretty bool
}

func NewJson() *Json {
	return &Json{pretty: false}
}
func NewJsonPretty() *Json {
	return &Json{pretty: true}
}

type BinaryData struct {
	data        []byte
	contentType string
	fileName    string
}

func NewBinaryData(data []byte, fileName, mimeType string) *BinaryData {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return &BinaryData{
		data:        data,
		fileName:    fileName,
		contentType: mimeType,
	}
}

func (j *Json) Write(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if bt, ok := data.(*BinaryData); ok {
		w.Header().Set("Content-Type", bt.contentType)
		if len(bt.fileName) > 0 {
			w.Header().Set(
				"Content-Disposition",
				`attachment; filename="`+html.EscapeString(bt.fileName)+`"`,
			)
		}
		w.WriteHeader(statusCode)
		_, err := w.Write(bt.data)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	enc := json.NewEncoder(w)
	if j.pretty {
		enc.SetIndent("", "    ")
	}

	return enc.Encode(data)
}

func (j *Json) Pretty() *Json {
	j.pretty = true
	return j
}
