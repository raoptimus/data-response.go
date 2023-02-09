package formatter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Binary struct{}

func NewBinary() *Binary {
	return &Binary{}
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// NewRequest - создает новую структуру запроса.
func NewRequest(contentType, fileName string, data []byte) *request {
	return &request{
		contentType: contentType,
		fileName:    fileName,
		data:        data,
	}
}

// request - структура запроса.
type request struct {
	contentType string
	fileName    string
	data        []byte
}

func (x *Binary) Write(w http.ResponseWriter, statusCode int, data any) error {
	req, ok := data.(*request)
	if statusCode == http.StatusOK && !ok {
		statusCode = http.StatusInternalServerError
		data = struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Name    string `json:"name"`
		}{
			Code:    statusCode,
			Message: "formatter: invalid request",
			Name:    "InternalServerError",
		}
	}

	if statusCode != http.StatusOK {
		w.Header().Set("Content-Type", contentTypeJson)
		w.WriteHeader(statusCode)
		return json.NewEncoder(w).Encode(data)
	}

	w.Header().Set("Content-Type", req.contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`form-data; name="field"; filename="%s"`, escapeQuotes(req.fileName)))
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(req.data)
	if err != nil {
		return err
	}
	return nil
}
