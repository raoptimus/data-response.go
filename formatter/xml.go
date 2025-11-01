package formatter

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

type Xml struct{}

func NewXml() *Xml {
	return &Xml{}
}

func (x *Xml) Marshal(header http.Header, data any) ([]byte, error) {
	bytes, err := xml.Marshal(data)
	if err != nil {
		return nil, errors.Wrapf(err, "encoding data '%T' to xml", data)
	}

	header.Set("Content-Type", "application/xml")

	return bytes, nil
}
