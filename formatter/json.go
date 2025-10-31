package formatter

import (
	json "github.com/json-iterator/go"
	"html"
	"net/http"
)

type Json struct {
	encoder json.API
	pretty  bool
}

func NewJson() *Json {
	return &Json{
		encoder: json.ConfigCompatibleWithStandardLibrary,
		pretty:  false,
	}
}
func NewJsonPretty() *Json {
	return NewJson().Pretty()
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
	header := w.Header()
	header.Set("X-Content-Type-Options", "nosniff")

	if bt, ok := data.(*BinaryData); ok {
		if _, err := w.Write(bt.data); err != nil {
			return err
		}

		header.Set("Content-Type", bt.contentType)
		if len(bt.fileName) > 0 {
			header.Set(
				"Content-Disposition",
				`attachment; filename="`+html.EscapeString(bt.fileName)+`"`,
			)
		}

		w.WriteHeader(statusCode)

		return nil
	}

	enc := j.encoder.NewEncoder(w)
	if j.pretty {
		enc.SetIndent("", "    ")
	}

	if err := enc.Encode(data); err != nil {
		return err
	}

	header.Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return nil
}

func (j *Json) Pretty() *Json {
	j.pretty = true
	return j
}
