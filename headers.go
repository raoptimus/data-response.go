package dataresponse

// HTTP Header names as constants to avoid typos and improve code clarity.
const (

	// General Headers

	HeaderCacheControl     = "Cache-Control"
	HeaderConnection       = "Connection"
	HeaderDate             = "Date"
	HeaderPragma           = "Pragma"
	HeaderTrailer          = "Trailer"
	HeaderTransferEncoding = "Transfer-Encoding"
	HeaderUpgrade          = "Upgrade"
	HeaderVia              = "Via"
	HeaderWarning          = "Warning"

	// Request Headers

	HeaderAccept             = "Accept"
	HeaderAcceptCharset      = "Accept-Charset"
	HeaderAcceptEncoding     = "Accept-Encoding"
	HeaderAcceptLanguage     = "Accept-Language"
	HeaderAuthorization      = "Authorization"
	HeaderExpect             = "Expect"
	HeaderFrom               = "From"
	HeaderHost               = "Host"
	HeaderIfMatch            = "If-Match"
	HeaderIfModifiedSince    = "If-Modified-Since"
	HeaderIfNoneMatch        = "If-None-Match"
	HeaderIfRange            = "If-Range"
	HeaderIfUnmodifiedSince  = "If-Unmodified-Since"
	HeaderMaxForwards        = "Max-Forwards"
	HeaderProxyAuthorization = "Proxy-Authorization"
	HeaderRange              = "Range"
	HeaderReferer            = "Referer"
	HeaderTE                 = "TE"
	HeaderUserAgent          = "User-Agent"

	// Response Headers

	HeaderAcceptRanges      = "Accept-Ranges"
	HeaderAge               = "Age"
	HeaderETag              = "ETag"
	HeaderLocation          = "Location"
	HeaderProxyAuthenticate = "Proxy-Authenticate"
	HeaderRetryAfter        = "Retry-After"
	HeaderServer            = "Server"
	HeaderVary              = "Vary"
	HeaderWWWAuthenticate   = "WWW-Authenticate"

	// Entity Headers

	HeaderAllow           = "Allow"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLanguage = "Content-Language"
	HeaderContentLength   = "Content-Length"
	HeaderContentLocation = "Content-Location"
	HeaderContentMD5      = "Content-MD5"
	HeaderContentRange    = "Content-Range"
	HeaderContentType     = "Content-Type"
	HeaderExpires         = "Expires"
	HeaderLastModified    = "Last-Modified"

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
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderOrigin                        = "Origin"

	// Security Headers

	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderReferrerPolicy          = "Referrer-Policy"
	HeaderPermissionsPolicy       = "Permissions-Policy"

	// Custom Common Headers

	HeaderXRequestID          = "X-Request-ID"
	HeaderXCorrelationID      = "X-Correlation-ID"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedHost      = "X-Forwarded-Host"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXCSRFToken          = "X-CSRF-Token"
	HeaderXRateLimitLimit     = "X-RateLimit-Limit"
	HeaderXRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderXRateLimitReset     = "X-RateLimit-Reset"

	// Content Disposition

	HeaderContentDisposition = "Content-Disposition"

	// API Versioning

	HeaderAPIVersion  = "API-Version"
	HeaderXAPIVersion = "X-API-Version"

	// Link Header (RFC 5988)

	HeaderLink = "Link"
)
