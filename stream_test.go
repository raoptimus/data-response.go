/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dataresponse "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/formatter"
	"github.com/raoptimus/data-response.go/v2/mockdataresponse"
	"github.com/raoptimus/data-response.go/v2/response"
)

const (
	// streamPattern is 16 bytes, so n repetitions give an exact size: 512 is 8 KiB, 65536 is 1 MiB.
	streamPattern = "0123456789abcdef"

	streamTestTimeout = 5 * time.Second
)

var (
	errStreamSource = errors.New("source failed")
	errStreamClient = errors.New("client gone")
	errStreamFormat = errors.New("format failed")
)

// =============================================================================
// Test helpers
// =============================================================================

type streamEvent struct {
	phase  response.StreamPhase
	err    error
	ctxErr error
}

// streamEvents collects observer calls; the channel makes them safe to read from the test goroutine.
type streamEvents struct {
	ch chan streamEvent
}

func newStreamEvents() *streamEvents {
	return &streamEvents{ch: make(chan streamEvent, 4)}
}

func (e *streamEvents) observe(ctx context.Context, phase response.StreamPhase, err error) {
	e.ch <- streamEvent{phase: phase, err: err, ctxErr: ctx.Err()}
}

func (e *streamEvents) next(t *testing.T) streamEvent {
	t.Helper()

	select {
	case ev := <-e.ch:
		return ev
	case <-time.After(streamTestTimeout):
		require.FailNow(t, "stream observer was not called")

		return streamEvent{}
	}
}

func (e *streamEvents) requireNone(t *testing.T) {
	t.Helper()

	select {
	case ev := <-e.ch:
		require.FailNowf(t, "unexpected stream observer call", "phase %s", ev.phase)
	default:
	}
}

// spyWriter is a http.ResponseWriter that records how the status and the body reach it.
type spyWriter struct {
	header      http.Header
	status      int
	headerCalls int
	writeCalls  int
	writeErr    error
	body        bytes.Buffer
}

func newSpyWriter() *spyWriter {
	return &spyWriter{header: make(http.Header)}
}

func (s *spyWriter) Header() http.Header {
	return s.header
}

func (s *spyWriter) WriteHeader(code int) {
	s.headerCalls++
	if s.status == 0 {
		s.status = code
	}
}

func (s *spyWriter) Write(p []byte) (int, error) {
	if s.status == 0 {
		s.WriteHeader(http.StatusOK)
	}

	s.writeCalls++
	if s.writeErr != nil {
		return 0, s.writeErr
	}

	return s.body.Write(p)
}

type failingFormatter struct {
	response.BaseFormatter
}

func (failingFormatter) Format(*response.DataResponse) (response.FormattedResponse, error) {
	return response.FormattedResponse{}, errStreamFormat
}

func (failingFormatter) ContentType() string {
	return response.ContentTypePlain
}

func newStreamFactory(events *streamEvents, opts ...dataresponse.Option) *dataresponse.Factory {
	base := []dataresponse.Option{dataresponse.WithFormatter(formatter.NewJSON())}
	if events != nil {
		base = append(base, dataresponse.WithStreamObserver(events.observe))
	}

	return dataresponse.New(append(base, opts...)...)
}

func streamHandler(fn func(ctx context.Context, w io.Writer) error) dataresponse.HandlerFunc {
	return func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
		return f.Stream(r.Context(), response.ContentTypeJSON, fn).
			WithHeader(response.HeaderXContentTypeOptions, response.ContentTypeOptionsNoSniff)
	}
}

func writePieces(w io.Writer, pieces []string) error {
	for _, piece := range pieces {
		if _, err := io.WriteString(w, piece); err != nil {
			return err
		}
	}

	return nil
}

func repeatPiece(piece string, n int) []string {
	pieces := make([]string, n)
	for i := range pieces {
		pieces[i] = piece
	}

	return pieces
}

func newStreamRequest() *http.Request {
	return httptest.NewRequest(http.MethodPost, "/", nil)
}

func postStream(t *testing.T, client *http.Client, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// =============================================================================
// Success: body and headers
// =============================================================================

func TestWrapHandler_Stream_WritesBodyAndHeaders_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		pieces       []string
		expectedBody string
	}{
		{
			name:         "single write",
			pieces:       []string{`[1,2,3]`},
			expectedBody: `[1,2,3]`,
		},
		{
			name:         "envelope and elements written by the caller",
			pieces:       []string{`{"data":[`, `{"x":1}`, `,{"x":2}`, `]}`},
			expectedBody: `{"data":[{"x":1},{"x":2}]}`,
		},
		{
			name:         "empty writes between data",
			pieces:       []string{"", "ab", "", "cd", ""},
			expectedBody: "abcd",
		},
		{
			name:         "exactly 8 KiB in one write",
			pieces:       []string{strings.Repeat(streamPattern, 512)},
			expectedBody: strings.Repeat(streamPattern, 512),
		},
		{
			name:         "8 KiB plus one byte in one write",
			pieces:       []string{strings.Repeat(streamPattern, 512) + "x"},
			expectedBody: strings.Repeat(streamPattern, 512) + "x",
		},
		{
			name:         "1 MiB in 4 KiB writes",
			pieces:       repeatPiece(strings.Repeat(streamPattern, 256), 256),
			expectedBody: strings.Repeat(streamPattern, 65536),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := newStreamEvents()
			h := dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
				return writePieces(w, tt.pieces)
			}), newStreamFactory(events))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, newStreamRequest())

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, len(tt.expectedBody), rec.Body.Len())
			assert.True(t, tt.expectedBody == rec.Body.String(), "body differs from the written pieces")
			assert.Equal(t, []string{response.ContentTypeJSON}, rec.Header().Values(response.HeaderContentType))
			assert.Equal(t, response.ContentTypeOptionsNoSniff, rec.Header().Get(response.HeaderXContentTypeOptions))
			assert.Empty(t, rec.Header().Get(response.HeaderContentLength))
			events.requireNone(t)
		})
	}
}

func TestWrapHandler_Stream_DeliversBodyOverHTTP_Successfully(t *testing.T) {
	tests := []struct {
		name                  string
		pieces                []string
		expectedBody          string
		expectedContentLength int64
	}{
		{
			name:                  "small body is sent with Content-Length",
			pieces:                []string{`[1,2,3]`},
			expectedBody:          `[1,2,3]`,
			expectedContentLength: 7,
		},
		{
			name:                  "exactly 8 KiB is chunked",
			pieces:                []string{strings.Repeat(streamPattern, 512)},
			expectedBody:          strings.Repeat(streamPattern, 512),
			expectedContentLength: -1,
		},
		{
			name:                  "8 KiB plus one byte is chunked",
			pieces:                []string{strings.Repeat(streamPattern, 512) + "x"},
			expectedBody:          strings.Repeat(streamPattern, 512) + "x",
			expectedContentLength: -1,
		},
		{
			name:                  "1 MiB in 4 KiB writes",
			pieces:                repeatPiece(strings.Repeat(streamPattern, 256), 256),
			expectedBody:          strings.Repeat(streamPattern, 65536),
			expectedContentLength: -1,
		},
		{
			name:                  "3 MiB in one write",
			pieces:                []string{strings.Repeat(streamPattern, 196608)},
			expectedBody:          strings.Repeat(streamPattern, 196608),
			expectedContentLength: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := newStreamEvents()
			srv := httptest.NewServer(dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
				return writePieces(w, tt.pieces)
			}), newStreamFactory(events)))
			defer srv.Close()

			resp := postStream(t, srv.Client(), srv.URL)
			defer resp.Body.Close()

			got, err := io.ReadAll(resp.Body)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expectedContentLength, resp.ContentLength)
			assert.Equal(t, len(tt.expectedBody), len(got))
			assert.True(t, tt.expectedBody == string(got), "body differs from the written pieces")
			events.requireNone(t)
		})
	}
}

func TestWrapHandler_Stream_ConcurrentRequests_Successfully(t *testing.T) {
	const clients = 16

	f := newStreamFactory(nil)
	srv := httptest.NewServer(dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
		return writePieces(w, repeatPiece(strings.Repeat(streamPattern, 256), 64))
	}), f))
	defer srv.Close()

	sizes := make([]int, clients)
	errs := make([]error, clients)

	var wg sync.WaitGroup
	for i := range clients {
		wg.Add(1)

		go func() {
			defer wg.Done()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL, nil)
			if err != nil {
				errs[i] = err

				return
			}

			resp, err := srv.Client().Do(req)
			if err != nil {
				errs[i] = err

				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			sizes[i], errs[i] = len(body), err
		}()
	}

	wg.Wait()

	for i := range clients {
		require.NoError(t, errs[i])
		assert.Equal(t, 262144, sizes[i])
	}
}

// =============================================================================
// Lazy commit of status and headers
// =============================================================================

func TestWrapHandler_Stream_SendsStatusAndHeadersWithFirstNonEmptyWrite_Successfully(t *testing.T) {
	spy := newSpyWriter()

	var (
		statusBefore, statusAfterEmptyWrite, statusAfterData int
		headerBefore                                         string
	)

	h := dataresponse.WrapHandler(dataresponse.HandlerFunc(
		func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
			return f.Stream(r.Context(), response.ContentTypeJSON, func(_ context.Context, w io.Writer) error {
				statusBefore = spy.status
				headerBefore = spy.Header().Get("X-Test")

				if _, err := w.Write(nil); err != nil {
					return err
				}
				statusAfterEmptyWrite = spy.status

				if _, err := io.WriteString(w, "data"); err != nil {
					return err
				}
				statusAfterData = spy.status

				return nil
			}).WithStatusCode(http.StatusCreated).WithHeader("X-Test", "1")
		}), newStreamFactory(nil))

	h.ServeHTTP(spy, newStreamRequest())

	assert.Equal(t, 0, statusBefore)
	assert.Empty(t, headerBefore)
	assert.Equal(t, 0, statusAfterEmptyWrite)
	assert.Equal(t, http.StatusCreated, statusAfterData)
	assert.Equal(t, "1", spy.Header().Get("X-Test"))
	assert.Equal(t, 1, spy.headerCalls)
	assert.Equal(t, "data", spy.body.String())
}

func TestWrapHandler_Stream_EmptyBodyStillSendsStatusAndHeaders_Successfully(t *testing.T) {
	tests := []struct {
		name string
		fn   func(ctx context.Context, w io.Writer) error
	}{
		{
			name: "no writes",
			fn:   func(context.Context, io.Writer) error { return nil },
		},
		{
			name: "only empty writes",
			fn: func(_ context.Context, w io.Writer) error {
				_, err := w.Write(nil)
				if err != nil {
					return err
				}
				_, err = io.WriteString(w, "")

				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := newStreamEvents()
			spy := newSpyWriter()
			h := dataresponse.WrapHandler(dataresponse.HandlerFunc(
				func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
					return f.Stream(r.Context(), response.ContentTypeJSON, tt.fn).
						WithStatusCode(http.StatusAccepted).
						WithHeader("X-Test", "1")
				}), newStreamFactory(events))

			h.ServeHTTP(spy, newStreamRequest())

			assert.Equal(t, http.StatusAccepted, spy.status)
			assert.Equal(t, 1, spy.headerCalls)
			assert.Equal(t, "1", spy.Header().Get("X-Test"))
			assert.Equal(t, []string{response.ContentTypeJSON}, spy.Header().Values(response.HeaderContentType))
			assert.Zero(t, spy.body.Len())
			events.requireNone(t)
		})
	}
}

// =============================================================================
// Failure before the first write
// =============================================================================

func TestWrapHandler_Stream_ErrorBeforeFirstWrite_RespondsInternalError_Failure(t *testing.T) {
	events := newStreamEvents()
	h := dataresponse.WrapHandler(
		streamHandler(func(context.Context, io.Writer) error { return errStreamSource }),
		newStreamFactory(events),
	)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, []string{response.ContentTypeJSON}, rec.Header().Values(response.HeaderContentType))
	assert.Empty(t, rec.Header().Get(response.HeaderXContentTypeOptions), "stream headers must not reach the error response")
	assert.Equal(t, "INTERNAL_SERVER_ERROR", body["code"])
	assert.Equal(t, "500", body["status"])

	ev := events.next(t)
	assert.Equal(t, response.StreamPhaseBeforeWrite, ev.phase)
	assert.ErrorIs(t, ev.err, errStreamSource)
	assert.NoError(t, ev.ctxErr)

	var streamErr *response.StreamError
	require.ErrorAs(t, ev.err, &streamErr)
	assert.NoError(t, streamErr.WriteErr)
	events.requireNone(t)
}

func TestWrapHandler_Stream_ErrorBeforeFirstWrite_CustomResponder_Failure(t *testing.T) {
	events := newStreamEvents()
	errFactory := dataresponse.New(dataresponse.WithFormatter(formatter.NewJSON()))

	var responderErr error
	f := newStreamFactory(events, dataresponse.WithStreamErrorResponder(
		func(ctx context.Context, err error) *response.DataResponse {
			responderErr = err

			return errFactory.ServiceUnavailable(ctx, "source unavailable")
		}))

	h := dataresponse.WrapHandler(streamHandler(func(context.Context, io.Writer) error { return errStreamSource }), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(t, `{"code":"SERVICE_UNAVAILABLE","status":"503","title":"source unavailable"}`, rec.Body.String())
	assert.Equal(t, errStreamSource, responderErr, "responder must get the error of the stream function itself")
	assert.Equal(t, response.StreamPhaseBeforeWrite, events.next(t).phase)
}

func TestWrapHandler_Stream_ErrorBeforeFirstWrite_ResponderReturnsNil_FallsBackToInternalError_Failure(t *testing.T) {
	f := newStreamFactory(nil, dataresponse.WithStreamErrorResponder(
		func(context.Context, error) *response.DataResponse { return nil }))

	h := dataresponse.WrapHandler(streamHandler(func(context.Context, io.Writer) error { return errStreamSource }), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", body["code"])
	assert.Equal(t, "Internal server error", body["title"])
}

func TestWrapHandler_Stream_ErrorBeforeFirstWrite_ObserverPanics_StillRespondsError_Failure(t *testing.T) {
	logger := mockdataresponse.NewLogger(t)
	logger.EXPECT().Error(mock.Anything, "stream observer panicked", "panic", "observer boom").Return()
	logger.EXPECT().Error(mock.Anything, "internal server error", "error", "source failed").Return()

	f := newStreamFactory(nil,
		dataresponse.WithLogger(logger),
		dataresponse.WithStreamObserver(func(context.Context, response.StreamPhase, error) { panic("observer boom") }),
	)
	h := dataresponse.WrapHandler(streamHandler(func(context.Context, io.Writer) error { return errStreamSource }), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestWrapHandler_Stream_ErrorBeforeFirstWrite_NoObserver_LogsOnlyInternalError_Failure(t *testing.T) {
	logger := mockdataresponse.NewLogger(t)
	logger.EXPECT().Error(mock.Anything, "internal server error", "error", "source failed").Return()

	f := newStreamFactory(nil, dataresponse.WithLogger(logger))
	h := dataresponse.WrapHandler(streamHandler(func(context.Context, io.Writer) error { return errStreamSource }), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestWrapHandler_Stream_ErrorBeforeFirstWrite_ErrorResponseFails_WritesPlainText_Failure(t *testing.T) {
	logger := mockdataresponse.NewLogger(t)
	logger.EXPECT().Error(mock.Anything, "failed to write error response", "error", "format failed").Return()

	badFactory := dataresponse.New(dataresponse.WithFormatter(failingFormatter{}))
	f := newStreamFactory(nil,
		dataresponse.WithLogger(logger),
		dataresponse.WithStreamErrorResponder(func(ctx context.Context, _ error) *response.DataResponse {
			return badFactory.Success(ctx, nil)
		}),
	)
	h := dataresponse.WrapHandler(streamHandler(func(context.Context, io.Writer) error { return errStreamSource }), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "Internal Server Error", rec.Body.String())
	assert.Equal(t, response.MimeTypePlainText.String(), rec.Header().Get(response.HeaderContentType))
}

func TestWrapHandler_Stream_PanicBeforeFirstWrite_Propagates_Failure(t *testing.T) {
	events := newStreamEvents()
	spy := newSpyWriter()
	h := dataresponse.WrapHandler(
		streamHandler(func(context.Context, io.Writer) error { panic("boom") }),
		newStreamFactory(events),
	)

	assert.PanicsWithValue(t, "boom", func() {
		h.ServeHTTP(spy, newStreamRequest())
	})
	assert.Equal(t, 0, spy.headerCalls)
	events.requireNone(t)
}

// =============================================================================
// Failure after the first write
// =============================================================================

func TestWrapHandler_Stream_ErrorAfterFirstWrite_AbortsHandler_Failure(t *testing.T) {
	events := newStreamEvents()
	spy := newSpyWriter()
	h := dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "part"); err != nil {
			return err
		}

		return errStreamSource
	}), newStreamFactory(events))

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		h.ServeHTTP(spy, newStreamRequest())
	})

	assert.Equal(t, http.StatusOK, spy.status)
	assert.Equal(t, 1, spy.headerCalls, "no error response may follow the started one")
	assert.Equal(t, "part", spy.body.String())

	ev := events.next(t)
	assert.Equal(t, response.StreamPhaseAfterWrite, ev.phase)
	assert.ErrorIs(t, ev.err, errStreamSource)

	var streamErr *response.StreamError
	require.ErrorAs(t, ev.err, &streamErr)
	assert.NoError(t, streamErr.WriteErr)
}

func TestWrapHandler_Stream_PanicAfterFirstWrite_AbortsHandlerAndReportsStack_Failure(t *testing.T) {
	events := newStreamEvents()
	spy := newSpyWriter()
	h := dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "part"); err != nil {
			return err
		}

		panic("boom")
	}), newStreamFactory(events))

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		h.ServeHTTP(spy, newStreamRequest())
	})

	assert.Equal(t, 1, spy.headerCalls)
	assert.Equal(t, "part", spy.body.String())

	ev := events.next(t)
	assert.Equal(t, response.StreamPhaseAfterWrite, ev.phase)

	var panicErr *response.StreamPanicError
	require.ErrorAs(t, ev.err, &panicErr)
	assert.Equal(t, "boom", panicErr.Value)
	assert.Contains(t, string(panicErr.Stack), "stream_test.go")
}

func TestWrapHandler_Stream_ErrorAfterFirstWrite_ObserverPanics_StillAbortsHandler_Failure(t *testing.T) {
	logger := mockdataresponse.NewLogger(t)
	logger.EXPECT().Error(mock.Anything, "stream observer panicked", "panic", "observer boom").Return()

	f := newStreamFactory(nil,
		dataresponse.WithLogger(logger),
		dataresponse.WithStreamObserver(func(context.Context, response.StreamPhase, error) { panic("observer boom") }),
	)
	spy := newSpyWriter()
	h := dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "part"); err != nil {
			return err
		}

		return errStreamSource
	}), f)

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		h.ServeHTTP(spy, newStreamRequest())
	})
	assert.Equal(t, 1, spy.headerCalls)
}

func TestWrapHandler_Stream_ErrorAfterFirstWrite_NoObserver_LogsAbort_Failure(t *testing.T) {
	logger := mockdataresponse.NewLogger(t)
	logger.EXPECT().Error(mock.Anything, "stream aborted", "error", "stream after_write: source failed").Return()

	f := newStreamFactory(nil, dataresponse.WithLogger(logger))
	spy := newSpyWriter()
	h := dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "part"); err != nil {
			return err
		}

		return errStreamSource
	}), f)

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		h.ServeHTTP(spy, newStreamRequest())
	})
}

// =============================================================================
// Client write errors and cancellation
// =============================================================================

func TestWrapHandler_Stream_ClientWriteFails_AbortsAndReportsWriteError_Failure(t *testing.T) {
	tests := []struct {
		name          string
		fn            func(ctx context.Context, w io.Writer) error
		expectedCause error
	}{
		{
			name: "write error returned as is",
			fn: func(_ context.Context, w io.Writer) error {
				_, err := w.Write([]byte("x"))

				return err
			},
			expectedCause: errStreamClient,
		},
		{
			name: "write error wrapped by the stream function",
			fn: func(_ context.Context, w io.Writer) error {
				if _, err := w.Write([]byte("x")); err != nil {
					return fmt.Errorf("%w: %w", errStreamSource, err)
				}

				return nil
			},
			expectedCause: errStreamSource,
		},
		{
			name: "write error ignored by the stream function",
			fn: func(_ context.Context, w io.Writer) error {
				_, _ = w.Write([]byte("x")) // the function deliberately drops the write error

				return errStreamSource
			},
			expectedCause: errStreamSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := newStreamEvents()
			spy := newSpyWriter()
			spy.writeErr = errStreamClient
			h := dataresponse.WrapHandler(streamHandler(tt.fn), newStreamFactory(events))

			assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
				h.ServeHTTP(spy, newStreamRequest())
			})

			assert.Equal(t, 1, spy.headerCalls, "no error response may follow the failed write")
			assert.Equal(t, 1, spy.writeCalls)

			ev := events.next(t)
			assert.Equal(t, response.StreamPhaseAfterWrite, ev.phase)
			assert.ErrorIs(t, ev.err, tt.expectedCause)

			var streamErr *response.StreamError
			require.ErrorAs(t, ev.err, &streamErr)
			assert.ErrorIs(t, streamErr.WriteErr, errStreamClient)
		})
	}
}

func TestWrapHandler_Stream_ContextCancelledMidStream_ReportsCancellation_Failure(t *testing.T) {
	events := newStreamEvents()
	spy := newSpyWriter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := dataresponse.WrapHandler(streamHandler(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "part"); err != nil {
			return err
		}
		cancel()

		return ctx.Err()
	}), newStreamFactory(events))

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		h.ServeHTTP(spy, newStreamRequest().WithContext(ctx))
	})

	ev := events.next(t)
	assert.Equal(t, response.StreamPhaseAfterWrite, ev.phase)
	assert.ErrorIs(t, ev.err, context.Canceled)
	assert.ErrorIs(t, ev.ctxErr, context.Canceled)
	assert.Equal(t, "part", spy.body.String())
}

func TestWrapHandler_Stream_ContextCancelledBeforeFirstWrite_RespondsWithResponder_Failure(t *testing.T) {
	events := newStreamEvents()
	errFactory := dataresponse.New(dataresponse.WithFormatter(formatter.NewJSON()))

	var responderErr error
	f := newStreamFactory(events, dataresponse.WithStreamErrorResponder(
		func(ctx context.Context, err error) *response.DataResponse {
			responderErr = err

			return errFactory.ServiceUnavailable(ctx, "cancelled")
		}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := dataresponse.WrapHandler(streamHandler(func(ctx context.Context, _ io.Writer) error {
		cancel()

		return ctx.Err()
	}), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest().WithContext(ctx))

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.ErrorIs(t, responderErr, context.Canceled)

	ev := events.next(t)
	assert.Equal(t, response.StreamPhaseBeforeWrite, ev.phase)
	assert.ErrorIs(t, ev.ctxErr, context.Canceled)
}

func TestWrapHandler_Stream_ClientDisconnects_StopsStreamAndReports_Failure(t *testing.T) {
	events := newStreamEvents()
	started := make(chan struct{})

	var once sync.Once

	srv := httptest.NewServer(dataresponse.WrapHandler(streamHandler(func(ctx context.Context, w io.Writer) error {
		chunk := []byte(strings.Repeat(streamPattern, 256))
		for {
			if err := ctx.Err(); err != nil {
				return err
			}

			if _, err := w.Write(chunk); err != nil {
				return err
			}

			once.Do(func() { close(started) })
		}
	}), newStreamFactory(events)))
	defer srv.Close()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL, nil)
	require.NoError(t, err)

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	select {
	case <-started:
	case <-time.After(streamTestTimeout):
		require.FailNow(t, "stream did not start")
	}

	_, err = io.ReadFull(resp.Body, make([]byte, 1024))
	require.NoError(t, err)

	cancel()

	ev := events.next(t)
	assert.Equal(t, response.StreamPhaseAfterWrite, ev.phase)

	var streamErr *response.StreamError
	require.ErrorAs(t, ev.err, &streamErr)
	assert.True(t, ev.ctxErr != nil || streamErr.WriteErr != nil, "client departure must show in ctx or in the write error")
}

// =============================================================================
// What the client sees after an abort
// =============================================================================

func TestWrapHandler_Stream_FailureAfterFirstWrite_ClientSeesAbortedBody_Failure(t *testing.T) {
	tests := []struct {
		name string
		fn   func(ctx context.Context, w io.Writer) error
	}{
		{
			name: "error",
			fn: func(_ context.Context, w io.Writer) error {
				if _, err := io.WriteString(w, strings.Repeat(streamPattern, 4096)); err != nil {
					return err
				}

				return errStreamSource
			},
		},
		{
			name: "panic",
			fn: func(_ context.Context, w io.Writer) error {
				if _, err := io.WriteString(w, strings.Repeat(streamPattern, 4096)); err != nil {
					return err
				}

				panic("boom")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := newStreamEvents()
			srv := httptest.NewServer(dataresponse.WrapHandler(streamHandler(tt.fn), newStreamFactory(events)))
			defer srv.Close()

			resp := postStream(t, srv.Client(), srv.URL)
			defer resp.Body.Close()

			_, err := io.ReadAll(resp.Body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
			assert.Equal(t, response.StreamPhaseAfterWrite, events.next(t).phase)
		})
	}
}

func TestWrapHandler_Stream_FailureAfterFirstWrite_HTTP2ClientSeesStreamError_Failure(t *testing.T) {
	events := newStreamEvents()
	srv := httptest.NewUnstartedServer(dataresponse.WrapHandler(streamHandler(func(_ context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, strings.Repeat(streamPattern, 4096)); err != nil {
			return err
		}

		return errStreamSource
	}), newStreamFactory(events)))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	defer srv.Close()

	resp := postStream(t, srv.Client(), srv.URL)
	defer resp.Body.Close()

	_, err := io.ReadAll(resp.Body)

	assert.Equal(t, 2, resp.ProtoMajor)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Error(t, err)
	assert.Equal(t, response.StreamPhaseAfterWrite, events.next(t).phase)
}

// =============================================================================
// Write
// =============================================================================

func TestWrite_Stream_ReturnsStreamError_Failure(t *testing.T) {
	tests := []struct {
		name                string
		writeErr            error
		fn                  func(w io.Writer) error
		expectedPhase       response.StreamPhase
		expectedCause       error
		expectedWriteErr    error
		expectedHeaderCalls int
	}{
		{
			name:                "error before first write",
			fn:                  func(io.Writer) error { return errStreamSource },
			expectedPhase:       response.StreamPhaseBeforeWrite,
			expectedCause:       errStreamSource,
			expectedWriteErr:    nil,
			expectedHeaderCalls: 0,
		},
		{
			name: "error after first write",
			fn: func(w io.Writer) error {
				if _, err := w.Write([]byte("x")); err != nil {
					return err
				}

				return errStreamSource
			},
			expectedPhase:       response.StreamPhaseAfterWrite,
			expectedCause:       errStreamSource,
			expectedWriteErr:    nil,
			expectedHeaderCalls: 1,
		},
		{
			name: "first write fails",
			fn: func(w io.Writer) error {
				_, err := w.Write([]byte("x"))

				return err
			},
			writeErr:            errStreamClient,
			expectedPhase:       response.StreamPhaseAfterWrite,
			expectedCause:       errStreamClient,
			expectedWriteErr:    errStreamClient,
			expectedHeaderCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := newSpyWriter()
			spy.writeErr = tt.writeErr
			resp := dataresponse.New().CreateDataResponse(http.StatusOK, nil).WithStream(tt.fn)

			err := dataresponse.Write(spy, resp)

			var streamErr *response.StreamError
			require.ErrorAs(t, err, &streamErr)
			assert.Equal(t, tt.expectedPhase, streamErr.Phase)
			assert.ErrorIs(t, err, tt.expectedCause)
			assert.Equal(t, tt.expectedWriteErr, streamErr.WriteErr)
			assert.Equal(t, tt.expectedHeaderCalls, spy.headerCalls)
		})
	}
}

// =============================================================================
// Factory.Stream
// =============================================================================

func TestFactory_Stream_SetsStatusAndContentType_Successfully(t *testing.T) {
	tests := []struct {
		name                string
		contentType         string
		expectedContentType string
	}{
		{name: "explicit content type", contentType: "application/x-ndjson", expectedContentType: "application/x-ndjson"},
		{name: "empty content type keeps the formatter one", contentType: "", expectedContentType: response.ContentTypeJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := dataresponse.New(dataresponse.WithFormatter(formatter.NewJSON()))

			resp := f.Stream(context.Background(), tt.contentType, func(context.Context, io.Writer) error { return nil })

			assert.True(t, resp.IsStream())
			assert.Equal(t, http.StatusOK, resp.StatusCode())
			assert.Equal(t, tt.expectedContentType, resp.ContentType())
			assert.Equal(t, []string{tt.expectedContentType}, resp.HeaderValues(response.HeaderContentType))
		})
	}
}

func TestFactory_Stream_PassesContextToStreamFunc_Successfully(t *testing.T) {
	type ctxKey struct{}

	ctx := context.WithValue(context.Background(), ctxKey{}, "marker")

	var got any
	resp := dataresponse.New().Stream(ctx, response.ContentTypeJSON, func(ctx context.Context, _ io.Writer) error {
		got = ctx.Value(ctxKey{})

		return nil
	})

	require.NoError(t, dataresponse.Write(httptest.NewRecorder(), resp))
	assert.Equal(t, "marker", got)
}

func TestFactory_Stream_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	logger := mockdataresponse.NewLogger(t)
	logger.EXPECT().Debug(mock.Anything, "stream response", "contentType", "application/x-ndjson").Return()

	f := dataresponse.New(dataresponse.WithLogger(logger), dataresponse.WithDebugMode(true))

	resp := f.Stream(context.Background(), "application/x-ndjson", func(context.Context, io.Writer) error { return nil })

	assert.True(t, resp.IsStream())
}

// =============================================================================
// Responses that are not streams
// =============================================================================

func TestWrapHandler_NonStreamResponses_AreWrittenAsBefore_Successfully(t *testing.T) {
	tests := []struct {
		name                  string
		handler               dataresponse.HandlerFunc
		expectedCode          int
		expectedBody          string
		expectedContentLength string
		expectedContentType   string
	}{
		{
			name: "success with data",
			handler: func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
				return f.Success(r.Context(), map[string]string{"k": "v"})
			},
			expectedCode:          http.StatusOK,
			expectedBody:          "{\"k\":\"v\"}\n",
			expectedContentLength: "10",
			expectedContentType:   response.ContentTypeJSON,
		},
		{
			name: "not found",
			handler: func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
				return f.NotFound(r.Context(), "no such user")
			},
			expectedCode:          http.StatusNotFound,
			expectedBody:          "{\"code\":\"NOT_FOUND\",\"status\":\"404\",\"title\":\"no such user\"}\n",
			expectedContentLength: "59",
			expectedContentType:   response.ContentTypeJSON,
		},
		{
			name: "binary with known size",
			handler: func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
				return f.Binary(r.Context(), io.NopCloser(strings.NewReader("hello")), "application/octet-stream", 5)
			},
			expectedCode:          http.StatusOK,
			expectedBody:          "hello",
			expectedContentLength: "5",
			expectedContentType:   "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := dataresponse.WrapHandler(tt.handler, newStreamFactory(newStreamEvents()))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, newStreamRequest())

			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
			assert.Equal(t, tt.expectedContentLength, rec.Header().Get(response.HeaderContentLength))
			assert.Equal(t, []string{tt.expectedContentType}, rec.Header().Values(response.HeaderContentType))
		})
	}
}

func TestWrapHandler_FormatterFails_RespondsInternalError_Failure(t *testing.T) {
	f := newStreamFactory(nil)
	h := dataresponse.WrapHandler(dataresponse.HandlerFunc(
		func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
			return f.Success(r.Context(), nil).WithFormatter(failingFormatter{})
		}), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, []string{response.ContentTypeJSON}, rec.Header().Values(response.HeaderContentType))
	assert.Equal(t, "INTERNAL_SERVER_ERROR", body["code"])
}

func TestWrapHandler_ErrorResponseFails_WritesPlainText_Failure(t *testing.T) {
	f := dataresponse.New(dataresponse.WithFormatter(failingFormatter{}))
	h := dataresponse.WrapHandler(dataresponse.HandlerFunc(
		func(r *http.Request, f *dataresponse.Factory) *response.DataResponse {
			return f.Success(r.Context(), nil)
		}), f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newStreamRequest())

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "Internal Server Error", rec.Body.String())
	assert.Equal(t, response.MimeTypePlainText.String(), rec.Header().Get(response.HeaderContentType))
}
