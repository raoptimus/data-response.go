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
)

// Write writes a DataResponse to http.ResponseWriter.
// It handles formatting, headers, and body writing.
func Write(w http.ResponseWriter, resp DataResponse) error {
	formattedResp, err := resp.Body()
	if err != nil {
		return err
	}
	if formattedResp.StreamSize > 0 {
		w.Header().Add(HeaderContentLength, strconv.FormatInt(formattedResp.StreamSize, 10))
	}

	// Write custom headers from response
	for key, values := range resp.Header() {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Binary-specific headers
	if resp.Filename() != "" {
		w.Header().Set(HeaderContentDisposition, `attachment; filename="`+resp.Filename()+`"`)
	}

	// Write status code
	w.WriteHeader(resp.StatusCode())

	// Stream data
	if formattedResp.StreamSize > 0 {
		_, err = io.CopyN(w, formattedResp.Stream, formattedResp.StreamSize)

		return err
	}

	_, err = io.Copy(w, formattedResp.Stream)

	return err
}
