package formatter

import (
	"bytes"
	"html"
	"net/http"

	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
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

func NewBinaryData(data []byte, fileName, mimeType string) BinaryData {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return BinaryData{
		data:        data,
		fileName:    fileName,
		contentType: mimeType,
	}
}

func (j *Json) Marshal(header http.Header, data any) ([]byte, error) {
	header.Set("X-Content-Type-Options", "nosniff")

	if bt, ok := data.(BinaryData); ok {
		header.Set("Content-Type", bt.contentType)
		if len(bt.fileName) > 0 {
			header.Set(
				"Content-Disposition",
				`attachment; filename="`+html.EscapeString(bt.fileName)+`"`,
			)
		}

		return bt.data, nil
	}

	var buffer bytes.Buffer
	enc := j.encoder.NewEncoder(&buffer)
	if j.pretty {
		enc.SetIndent("", "    ")
	}

	if err := enc.Encode(data); err != nil {
		return nil, errors.Wrapf(err, "encoding data '%T' to json", data)
	}

	header.Set("Content-Type", "application/json")

	return buffer.Bytes(), nil
}

func (j *Json) Pretty() *Json {
	j.pretty = true
	return j
}
