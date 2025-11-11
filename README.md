# data-response

High-performance Go library for standardized HTTP responses with support for multiple formats, compression, and middleware.

## Features

✅ **Multiple Formatters** - JSON, XML, HTML, Binary  
✅ **Compression** - Gzip with different compression levels  
✅ **Content Negotiation** - Automatic format selection based on Accept header  
✅ **Logger Agnostic** - slog, logrus, zap, zerolog adapters  
✅ **HTTP Headers** - Full support for http.Header with multiple values  
✅ **Binary Support** - Efficient file streaming with Content-Length support  
✅ **Middleware** - Recovery, compression, content type validation  
✅ **Security** - Verbosity control for error details  
✅ **Production Ready** - Type-safe, well-tested, battle-ready  

```shell
go get module github.com/raoptimus/data-response.go
```

## Quick Start

```go
package main

import (
	"log/slog"
	"net/http"

	dr "module github.com/raoptimus/data-response.go"
	"module github.com/raoptimus/data-response.go/adapters"
	"module github.com/raoptimus/data-response.go/formatters"
)

func main() {
	factory := dr.New(
		dr.WithLogger(adapters.NewSlog(slog.Default())),
		dr.WithFormatter(formatters.NewJSON()),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := factory.Success(r.Context(), map[string]string{
			"message": "Hello, World!",
		})
		dr.Write(w, r, resp, factory)
	})

	http.ListenAndServe(":8080", nil)
}
```

## HTTP Headers Support

```go
resp := factory.Success(ctx, data)
// Add single header
resp = resp.WithHeader("X-Custom", "value")
// Add multiple values for same header
resp = resp.WithHeader("Set-Cookie", "session=abc")
resp = resp.WithHeader("Set-Cookie", "path=/")
// Get all values
values := resp.HeaderValues("Set-Cookie") // ["session=abc", "path=/"]
// Check if header exists
if resp.HasHeader("X-Custom") {
// ...
}

```

## HTTP Codes and MIME Types

```go
// Use HTTPCode constants
code := dr.HTTPCodeNotFound
// Get from status
code = dr.CodeFromStatus(http.StatusOK) // HTTPCodeOK
// Use MIME types
mimeType := dr.MimeTypeJSON
mimeType = dr.MimeTypeFromExtension(".pdf") // MimeTypePDF
```
