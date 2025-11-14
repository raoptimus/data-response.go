/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

// HTTP Header names as constants.
const (
	// General Headers

	HeaderCacheControl     = "Cache-Control"
	HeaderConnection       = "Connection"
	HeaderDate             = "Date"
	HeaderTransferEncoding = "Transfer-Encoding"

	// Request Headers

	HeaderAccept             = "Accept"
	HeaderAcceptEncoding     = "Accept-Encoding"
	HeaderAcceptLanguage     = "Accept-Language"
	HeaderAuthorization      = "Authorization"
	HeaderHost               = "Host"
	HeaderIfModifiedSince    = "If-Modified-Since"
	HeaderIfNoneMatch        = "If-None-Match"
	HeaderUserAgent          = "User-Agent"
	HeaderReferer = "Referer"

	// Response Headers

	HeaderETag            = "ETag"
	HeaderLocation        = "Location"
	HeaderRetryAfter      = "Retry-After"
	HeaderServer          = "Server"
	HeaderVary            = "Vary"
	HeaderWWWAuthenticate = "WWW-Authenticate"

	// Entity Headers

	HeaderContentEncoding    = "Content-Encoding"
	HeaderContentLanguage    = "Content-Language"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderLastModified       = "Last-Modified"

	// Cookie Headers

	HeaderCookie    = "Cookie"
	HeaderSetCookie = "Set-Cookie"

	// CORS Headers

	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderOrigin                        = "Origin"

	// Security Headers

	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderReferrerPolicy          = "Referrer-Policy"

	// Custom Headers

	HeaderXRequestID          = "X-Request-ID"
	HeaderXCorrelationID      = "X-Correlation-ID"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRateLimitLimit     = "X-RateLimit-Limit"
	HeaderXRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderXRateLimitReset     = "X-RateLimit-Reset"
	HeaderAPIVersion = "API-Version"
)

// Common header values as constants.
const (
	// Content-Type values

	ContentTypeJSON             = "application/json"
	ContentTypeJSONCharsetUTF8  = "application/json; charset=utf-8"
	ContentTypeXML              = "application/xml"
	ContentTypeXMLCharsetUTF8   = "application/xml; charset=utf-8"
	ContentTypeHTML             = "text/html"
	ContentTypeHTMLCharsetUTF8  = "text/html; charset=utf-8"
	ContentTypePlain            = "text/plain"
	ContentTypePlainCharsetUTF8 = "text/plain; charset=utf-8"
	ContentTypeForm             = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm    = "multipart/form-data"
	ContentTypeOctetStream      = "application/octet-stream"

	// Cache-Control values

	CacheControlNoCache = "no-cache"
	CacheControlNoStore = "no-store"
	CacheControlPublic  = "public"
	CacheControlPrivate = "private"
	CacheControlMaxAge  = "max-age"

	// Connection values

	ConnectionKeepAlive = "keep-alive"
	ConnectionClose     = "close"

	// Content-Encoding values

	ContentEncodingGzip    = "gzip"
	ContentEncodingDeflate = "deflate"
	ContentEncodingBrotli  = "br"

	// X-Content-Type-Options values

	ContentTypeOptionsNoSniff = "nosniff"

	// X-Frame-Options values

	FrameOptionsDeny       = "DENY"
	FrameOptionsSameOrigin = "SAMEORIGIN"

	// Referrer-Policy values

	ReferrerPolicyNoReferrer                  = "no-referrer"
	ReferrerPolicyStrictOriginWhenCrossOrigin = "strict-origin-when-cross-origin"
)
