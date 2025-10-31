package formatter

import (
	"encoding/xml"
	"net/http"
)

type Xml struct{}

func NewXml() *Xml {
	return &Xml{}
}

func (x *Xml) Write(w http.ResponseWriter, statusCode int, data any) error {
	if err := xml.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)

	return nil
}
