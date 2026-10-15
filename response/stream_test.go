/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package response_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raoptimus/data-response.go/v2/response"
)

var errStreamTest = errors.New("stream test error")

func TestStreamPhase_String_ReturnsName_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		phase    response.StreamPhase
		expected string
	}{
		{name: "before write", phase: response.StreamPhaseBeforeWrite, expected: "before_write"},
		{name: "after write", phase: response.StreamPhaseAfterWrite, expected: "after_write"},
		{name: "zero value", phase: 0, expected: "unknown"},
		{name: "value past the last phase", phase: response.StreamPhaseAfterWrite + 1, expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.String())
		})
	}
}

func TestStreamError_Error_ReturnsPhaseAndCause_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		err      *response.StreamError
		expected string
	}{
		{
			name:     "before write with cause",
			err:      &response.StreamError{Phase: response.StreamPhaseBeforeWrite, Err: errStreamTest},
			expected: "stream before_write: stream test error",
		},
		{
			name:     "after write with cause",
			err:      &response.StreamError{Phase: response.StreamPhaseAfterWrite, Err: errStreamTest},
			expected: "stream after_write: stream test error",
		},
		{
			name:     "without cause",
			err:      &response.StreamError{Phase: response.StreamPhaseAfterWrite},
			expected: "stream after_write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestStreamError_Unwrap_ReturnsCause_Successfully(t *testing.T) {
	err := error(&response.StreamError{Phase: response.StreamPhaseBeforeWrite, Err: errStreamTest})

	assert.ErrorIs(t, err, errStreamTest)
	assert.NoError(t, errors.Unwrap(&response.StreamError{}))
}

func TestStreamPanicError_ErrorAndUnwrap_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		value           any
		expectedMessage string
		expectedCause   error
	}{
		{name: "string value", value: "boom", expectedMessage: "panic: boom", expectedCause: nil},
		{name: "int value", value: 42, expectedMessage: "panic: 42", expectedCause: nil},
		{
			name:            "error value",
			value:           errStreamTest,
			expectedMessage: "panic: stream test error",
			expectedCause:   errStreamTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &response.StreamPanicError{Value: tt.value, Stack: []byte("stack")}

			assert.Equal(t, tt.expectedMessage, err.Error())
			assert.Equal(t, tt.expectedCause, err.Unwrap())
		})
	}
}

func TestDataResponse_WithStream_MarksResponseAsStream_Successfully(t *testing.T) {
	tests := []struct {
		name             string
		resp             *response.DataResponse
		expectedIsStream bool
		expectedFunc     bool
	}{
		{
			name:             "response with stream",
			resp:             response.NewDataResponse(http.StatusOK, nil).WithStream(func(io.Writer) error { return nil }),
			expectedIsStream: true,
			expectedFunc:     true,
		},
		{
			name:             "plain response",
			resp:             response.NewDataResponse(http.StatusOK, "data"),
			expectedIsStream: false,
			expectedFunc:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedIsStream, tt.resp.IsStream())
			assert.Equal(t, tt.expectedFunc, tt.resp.StreamFunc() != nil)
		})
	}
}

func TestDataResponse_WithStream_BodyHasUnknownSize_Successfully(t *testing.T) {
	resp := response.NewDataResponse(http.StatusOK, nil).WithStream(func(io.Writer) error { return nil })

	body, err := resp.Body()

	require.NoError(t, err)
	require.NotNil(t, body.Stream)
	assert.Equal(t, int64(-1), body.StreamSize)
}

func TestDataResponse_WithStream_BodyCopyRunsStreamFunc_Successfully(t *testing.T) {
	calls := 0
	resp := response.NewDataResponse(http.StatusOK, nil).WithStream(func(w io.Writer) error {
		calls++
		_, err := io.WriteString(w, "abc")

		return err
	})

	body, err := resp.Body()
	require.NoError(t, err)

	var buf bytes.Buffer
	n, err := io.Copy(&buf, body.Stream)

	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
	assert.Equal(t, "abc", buf.String())
	assert.Equal(t, 1, calls)
}

func TestDataResponse_WithStream_BodyCopyReturnsStreamError_Failure(t *testing.T) {
	resp := response.NewDataResponse(http.StatusOK, nil).WithStream(func(w io.Writer) error {
		if _, err := io.WriteString(w, "ab"); err != nil {
			return err
		}

		return errStreamTest
	})

	body, err := resp.Body()
	require.NoError(t, err)

	var buf bytes.Buffer
	n, err := io.Copy(&buf, body.Stream)

	assert.ErrorIs(t, err, errStreamTest)
	assert.Equal(t, int64(2), n)
	assert.Equal(t, "ab", buf.String())
}

func TestDataResponse_WithStream_BodyRead_Failure(t *testing.T) {
	resp := response.NewDataResponse(http.StatusOK, nil).WithStream(func(io.Writer) error { return nil })

	body, err := resp.Body()
	require.NoError(t, err)

	n, err := body.Stream.Read(make([]byte, 1))

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}
