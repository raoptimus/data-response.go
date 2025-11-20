/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package formatter

import (
	"bytes"
	"html/template"

	"github.com/raoptimus/data-response.go/v2/internal/conv"
	"github.com/raoptimus/data-response.go/v2/response"
)

// HTML is an HTML response formatter.
type HTML struct {
	response.BaseFormatter
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
func (f *HTML) Format(resp *response.DataResponse) (response.FormattedResponse, error) {
	if resp.IsBinary() {
		return response.FormattedResponse{}, response.NewError(errCode500, "cannot format binary as HTML")
	}

	var buf bytes.Buffer

	if f.template != nil {
		// Pass resp.Data() to template
		if err := f.template.Execute(&buf, resp.Data()); err != nil {
			return response.FormattedResponse{}, response.WrapError(errCode500, err, "failed to execute template")
		}
	} else {
		// Default template
		if err := f.defaultTemplate(&buf, resp); err != nil {
			return response.FormattedResponse{}, err
		}
	}

	return response.FormattedResponse{
		Stream:     bytes.NewReader(buf.Bytes()),
		StreamSize: int64(buf.Len()),
	}, nil
}

func (f *HTML) defaultTemplate(buf *bytes.Buffer, resp *response.DataResponse) error {
	dataBytes, err := conv.DataToString(resp.Data())
	if err != nil {
		return err
	}

	_, err = buf.Write(dataBytes)

	return err
}

// ContentType returns text/html.
func (f *HTML) ContentType() string {
	return response.ContentTypeHTML
}
