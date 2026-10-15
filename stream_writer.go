/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"net/http"
	"runtime/debug"

	"github.com/raoptimus/data-response.go/v2/response"
)

// lazyWriter holds back status and headers of the response until the first non-empty body write.
type lazyWriter struct {
	w         http.ResponseWriter
	resp      *response.DataResponse
	committed bool
	writeErr  error
}

func (l *lazyWriter) Write(p []byte) (int, error) {
	// Encoders flush empty buffers; that must not decide the status.
	if len(p) == 0 {
		return 0, nil
	}

	l.commit()

	n, err := l.w.Write(p)
	if err != nil && l.writeErr == nil {
		l.writeErr = err
	}

	return n, err
}

func (l *lazyWriter) commit() {
	if l.committed {
		return
	}

	l.committed = true

	headers := l.w.Header()
	for key, values := range l.resp.Header() {
		for _, value := range values {
			headers.Add(key, value)
		}
	}

	l.w.WriteHeader(l.resp.StatusCode())
}

func (l *lazyWriter) streamError(err error) *response.StreamError {
	phase := response.StreamPhaseBeforeWrite
	if l.committed {
		phase = response.StreamPhaseAfterWrite
	}

	return &response.StreamError{Phase: phase, Err: err, WriteErr: l.writeErr}
}

// writeStream runs the stream body. Status and headers reach the client with the first write only,
// so a failure before it can still be turned into a regular error response.
// A panic before the first write is not touched; after it the panic becomes a *response.StreamError.
func writeStream(w http.ResponseWriter, resp *response.DataResponse) (err error) {
	defer resp.Close()

	lw := &lazyWriter{w: w, resp: resp}

	defer func() {
		if !lw.committed {
			return
		}

		rcv := recover()
		if rcv == nil {
			return
		}

		err = lw.streamError(&response.StreamPanicError{Value: rcv, Stack: debug.Stack()})
	}()

	if fnErr := resp.StreamFunc()(lw); fnErr != nil {
		return lw.streamError(fnErr)
	}

	// An empty body still needs its status and headers.
	lw.commit()

	return nil
}
