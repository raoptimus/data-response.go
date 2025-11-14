/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import "net/http"

// HTTPCode represents a standardized HTTP error code.
type HTTPCode string

// HTTP Code constants for all standard HTTP status codes.
const (
	// 2xx Success codes

	HTTPCodeOK        HTTPCode = "OK"
	HTTPCodeCreated   HTTPCode = "CREATED"
	HTTPCodeAccepted  HTTPCode = "ACCEPTED"
	HTTPCodeNoContent HTTPCode = "NO_CONTENT"

	// 3xx Redirection codes

	HTTPCodeMovedPermanently  HTTPCode = "MOVED_PERMANENTLY"
	HTTPCodeFound             HTTPCode = "FOUND"
	HTTPCodeNotModified       HTTPCode = "NOT_MODIFIED"
	HTTPCodeTemporaryRedirect HTTPCode = "TEMPORARY_REDIRECT"
	HTTPCodePermanentRedirect HTTPCode = "PERMANENT_REDIRECT"

	// 4xx Client Error codes

	HTTPCodeBadRequest                  HTTPCode = "BAD_REQUEST"
	HTTPCodeUnauthorized                HTTPCode = "UNAUTHORIZED"
	HTTPCodeForbidden                   HTTPCode = "FORBIDDEN"
	HTTPCodeNotFound                    HTTPCode = "NOT_FOUND"
	HTTPCodeMethodNotAllowed            HTTPCode = "METHOD_NOT_ALLOWED"
	HTTPCodeConflict                    HTTPCode = "CONFLICT"
	HTTPCodeGone                        HTTPCode = "GONE"
	HTTPCodeLengthRequired              HTTPCode = "LENGTH_REQUIRED"
	HTTPCodePreconditionFailed          HTTPCode = "PRECONDITION_FAILED"
	HTTPCodePayloadTooLarge             HTTPCode = "PAYLOAD_TOO_LARGE"
	HTTPCodeURITooLong                  HTTPCode = "URI_TOO_LONG"
	HTTPCodeUnsupportedMediaType        HTTPCode = "UNSUPPORTED_MEDIA_TYPE"
	HTTPCodeUnprocessableEntity         HTTPCode = "UNPROCESSABLE_ENTITY"
	HTTPCodeLocked                      HTTPCode = "LOCKED"
	HTTPCodeTooEarly                    HTTPCode = "TOO_EARLY"
	HTTPCodeUpgradeRequired             HTTPCode = "UPGRADE_REQUIRED"
	HTTPCodePreconditionRequired        HTTPCode = "PRECONDITION_REQUIRED"
	HTTPCodeTooManyRequests             HTTPCode = "TOO_MANY_REQUESTS"
	HTTPCodeRequestHeaderFieldsTooLarge HTTPCode = "REQUEST_HEADER_FIELDS_TOO_LARGE"
	HTTPCodeUnavailableForLegalReasons  HTTPCode = "UNAVAILABLE_FOR_LEGAL_REASONS"

	// 5xx Server Error codes

	HTTPCodeInternalServerError           HTTPCode = "INTERNAL_SERVER_ERROR"
	HTTPCodeNotImplemented                HTTPCode = "NOT_IMPLEMENTED"
	HTTPCodeBadGateway                    HTTPCode = "BAD_GATEWAY"
	HTTPCodeServiceUnavailable            HTTPCode = "SERVICE_UNAVAILABLE"
	HTTPCodeGatewayTimeout                HTTPCode = "GATEWAY_TIMEOUT"
	HTTPCodeHTTPVersionNotSupported       HTTPCode = "HTTP_VERSION_NOT_SUPPORTED"
	HTTPCodeVariantAlsoNegotiates         HTTPCode = "VARIANT_ALSO_NEGOTIATES"
	HTTPCodeInsufficientStorage           HTTPCode = "INSUFFICIENT_STORAGE"
	HTTPCodeLoopDetected                  HTTPCode = "LOOP_DETECTED"
	HTTPCodeNotExtended                   HTTPCode = "NOT_EXTENDED"
	HTTPCodeNetworkAuthenticationRequired HTTPCode = "NETWORK_AUTHENTICATION_REQUIRED"
)

// String returns the string representation of HTTPCode.
func (c HTTPCode) String() string {
	return string(c)
}

// HTTPCodesMapping maps HTTP status codes to HTTPCode constants.
// It provides a bidirectional mapping for converting between numeric status codes
// and their string representations.
var HTTPCodesMapping = map[int]HTTPCode{
	// 2xx Success
	http.StatusOK:        HTTPCodeOK,
	http.StatusCreated:   HTTPCodeCreated,
	http.StatusAccepted:  HTTPCodeAccepted,
	http.StatusNoContent: HTTPCodeNoContent,

	// 3xx Redirection
	http.StatusMovedPermanently:  HTTPCodeMovedPermanently,
	http.StatusFound:             HTTPCodeFound,
	http.StatusNotModified:       HTTPCodeNotModified,
	http.StatusTemporaryRedirect: HTTPCodeTemporaryRedirect,
	http.StatusPermanentRedirect: HTTPCodePermanentRedirect,

	// 4xx Client Error
	http.StatusBadRequest:            HTTPCodeBadRequest,
	http.StatusUnauthorized:          HTTPCodeUnauthorized,
	http.StatusForbidden:             HTTPCodeForbidden,
	http.StatusNotFound:              HTTPCodeNotFound,
	http.StatusMethodNotAllowed:      HTTPCodeMethodNotAllowed,
	http.StatusConflict:              HTTPCodeConflict,
	http.StatusGone:                  HTTPCodeGone,
	http.StatusLengthRequired:        HTTPCodeLengthRequired,
	http.StatusPreconditionFailed:    HTTPCodePreconditionFailed,
	http.StatusRequestEntityTooLarge: HTTPCodePayloadTooLarge,
	http.StatusRequestURITooLong:     HTTPCodeURITooLong,
	http.StatusUnsupportedMediaType:  HTTPCodeUnsupportedMediaType,
	http.StatusUnprocessableEntity:   HTTPCodeUnprocessableEntity,
	http.StatusLocked:                HTTPCodeLocked,
	http.StatusTooEarly:              HTTPCodeTooEarly,
	http.StatusUpgradeRequired:       HTTPCodeUpgradeRequired,
	http.StatusPreconditionRequired:  HTTPCodePreconditionRequired,
	http.StatusTooManyRequests:       HTTPCodeTooManyRequests,

	// 5xx Server Error
	http.StatusInternalServerError:           HTTPCodeInternalServerError,
	http.StatusNotImplemented:                HTTPCodeNotImplemented,
	http.StatusBadGateway:                    HTTPCodeBadGateway,
	http.StatusServiceUnavailable:            HTTPCodeServiceUnavailable,
	http.StatusGatewayTimeout:                HTTPCodeGatewayTimeout,
	http.StatusHTTPVersionNotSupported:       HTTPCodeHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates:         HTTPCodeVariantAlsoNegotiates,
	http.StatusInsufficientStorage:           HTTPCodeInsufficientStorage,
	http.StatusLoopDetected:                  HTTPCodeLoopDetected,
	http.StatusNotExtended:                   HTTPCodeNotExtended,
	http.StatusNetworkAuthenticationRequired: HTTPCodeNetworkAuthenticationRequired,
}

// CodeFromStatus returns HTTPCode for given HTTP status code.
// If the status code is not recognized, it returns an empty HTTPCode.
func CodeFromStatus(status int) HTTPCode {
	if code, ok := HTTPCodesMapping[status]; ok {
		return code
	}

	return ""
}
