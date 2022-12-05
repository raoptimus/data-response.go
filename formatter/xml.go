package formatter

import (
	"encoding/xml"
	"net/http"
)

type Xml struct{}

func NewXml() *Xml {
	return &Xml{}
}

func (x *Xml) Write(data any, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/xml")
	return xml.NewEncoder(w).Encode(data)
}
