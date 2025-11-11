package formatter

import (
	"html/template"
	"net/http"

	"github.com/pkg/errors"
	dataresponse "github.com/raoptimus/data-response.go"
)

var ErrDataIsNotStringable = errors.New("data is not a string-able")

// HTML is an HTML response formatter.
type HTML struct {
	dataresponse.BaseFormatter
	template *template.Template
}

// NewHTML creates a new HTML formatter.
func NewHTML() *HTML {
	return &HTML{}
}

func (f *HTML) WithTemplate(tmpl *template.Template) *HTML {
	f.template = tmpl

	return f
}

// Format writes HTML response.
func (f *HTML) Format(w http.ResponseWriter, resp dataresponse.DataResponse) error {
	if resp.IsBinary() {
		return dataresponse.NewError(http.StatusInternalServerError, "cannot format binary as HTML")
	}

	f.WriteHeaders(w, resp, f.ContentType())
	w.WriteHeader(resp.StatusCode())

	if f.template != nil {
		return f.template.Execute(w, resp)
	}

	return f.defaultTemplate(w, resp)
}

func (f *HTML) defaultTemplate(w http.ResponseWriter, resp dataresponse.DataResponse) error {
	dataBytes, err := f.dataToString(resp.Data())
	if err != nil {
		return err
	}
	
	_, err = w.Write(dataBytes)

	return err
}

// ContentType returns text/html.
func (f *HTML) ContentType() string {
	return dataresponse.MimeTypeHTML.String()
}

func (f *HTML) dataToString(data any) ([]byte, error) {
	switch v := data.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, errors.WithStack(ErrDataIsNotStringable)
	}
}
