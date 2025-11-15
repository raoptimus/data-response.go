/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package formatter

import (
	"bufio"
	"bytes"
	"html/template"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/internal/conv"
)

// HTML is an HTML response formatter.
type HTML struct {
	dr.BaseFormatter
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
func (f *HTML) Format(resp dr.DataResponse) (dr.FormattedResponse, error) {
	if resp.IsBinary() {
		return dr.FormattedResponse{}, dr.NewError(500, "cannot format binary as HTML")
	}

	var buf bytes.Buffer

	if f.template != nil {
		// Pass resp.Data() to template
		if err := f.template.Execute(&buf, resp.Data()); err != nil {
			return dr.FormattedResponse{}, dr.WrapError(500, err, "failed to execute template")
		}
	} else {
		// Default template
		if err := f.defaultTemplate(&buf, resp); err != nil {
			return dr.FormattedResponse{}, err
		}
	}

	return dr.FormattedResponse{
		ContentType: f.ContentType(),
		Stream:     bufio.NewReader(&buf),
		StreamSize: int64(buf.Len()),
	}, nil
}

func (f *HTML) defaultTemplate(buf *bytes.Buffer, resp dr.DataResponse) error {
	dataBytes, err := conv.DataToString(resp.Data())
	if err != nil {
		return err
	}

	_, err = buf.Write(dataBytes)

	return err
}

// ContentType returns text/html.
func (f *HTML) ContentType() string {
	return dr.ContentTypeHTML
}

