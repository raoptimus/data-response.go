/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"io"
	"net/http"
	"strconv"

	"github.com/raoptimus/data-response.go/v2/response"
)

// Write writes a DataResponse to http.ResponseWriter.
// It handles formatting, headers, and body writing.
func Write(w http.ResponseWriter, resp *response.DataResponse) error {
	defer resp.Close()

	formattedResp, err := resp.Body()
	if err != nil {
		return err
	}

	headers := w.Header()
	if formattedResp.StreamSize > 0 {
		headers.Add(response.HeaderContentLength, strconv.FormatInt(formattedResp.StreamSize, 10))
	}

	// Write custom headers from response
	for key, values := range resp.Header() {
		for _, value := range values {
			headers.Add(key, value)
		}
	}

	// Binary-specific headers
	if resp.Filename() != "" {
		headers.Set(response.HeaderContentDisposition, `attachment; filename="`+resp.Filename()+`"`)
	}

	// Write status code
	w.WriteHeader(resp.StatusCode())

	// Stream data
	if formattedResp.StreamSize > 0 {
		_, err = io.CopyN(w, formattedResp.Stream, formattedResp.StreamSize)

		return err
	}

	if formattedResp.Stream != nil {
		_, err = io.Copy(w, formattedResp.Stream)
	}

	return err
}
