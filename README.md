# data-response.go

[![Go Reference](https://pkg.go.dev/badge/github.com/raoptimus/data-response.go.svg)](https://pkg.go.dev/github.com/raoptimus/data-response.go/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/raoptimus/data-response.go)](https://goreportcard.com/report/github.com/raoptimus/data-response.go)
[![License](https://img.shields.io/github/license/raoptimus/data-response.go)](https://https://github.com/raoptimus/data-response.go/blob/main/LICENSE)

A functional and composable Go library for building standardized HTTP responses with type-safe middleware support.

## Features

- 🎯 **Type-Safe Responses** - Immutable response pattern with fluent API
- 🔗 **Composable Middleware** - Functional middleware chain with DataResponse support
- 🔌 **Chi Router Integration** - Seamless chi middleware adapter
- 📝 **Flexible Logging** - Template-based access log formatting with context support
- 🔒 **Security Headers** - Built-in helpers for CORS, CSP, and security headers
- 📊 **Metrics Support** - Built-in measurement middleware with customizable metrics service
- 🎨 **Custom Formatters** - JSON, XML, or implement your own formatter
- 🗂️ **Binary Responses** - File serving and streaming support
- ⚡ **Error Handling** - Stack trace preservation with verbosity control
- 🧪 **Well Tested** - Comprehensive unit test coverage

## Installation

```shell
go get github.com/raoptimus/data-response.go/v2
```

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    dr "github.com/raoptimus/data-response.go/v2"
    "github.com/raoptimus/data-response.go/v2/formatter"
    "github.com/raoptimus/data-response.go/v2/middleware"
    "github.com/raoptimus/data-response.go/pkg/chiadapter"
)

func main() {
    // Create factory
    factory := dr.New(
        dr.WithFormatter(formatter.NewJSON()),
        dr.WithVerbosity(false), // Hide errors in production
    )

    // Create router with DataResponse support
    r := chiadapter.NewRouter(factory)
    
    // Add middleware
    r.WithMiddleware(
        middleware.RequestTimer(),
        middleware.Recovery(),
        middleware.LoggingDefault(),
    )
    
    // Define routes
    r.Get("/users/{id}", getUser)
    r.Post("/users", createUser)
    
    http.ListenAndServe(":8080", r)
}

func getUser(r *http.Request, f *dr.Factory) dr.DataResponse {
    // Your logic here
    user := map[string]interface{}{
        "id": chi.URLParam(r, "id"),
        "name": "John Doe",
    }
	
    return f.Success(r.Context(), user)
}

func createUser(r *http.Request, f *dr.Factory) dr.DataResponse {
    // Validation example
    if err := validateInput(r); err != nil {
        return f.BadRequest(r.Context(), "Invalid input")
    }

    // Create user logic...
    return f.Created(r.Context(), newUser, "/users/123")
}
```

## Core Concepts

### Factory Pattern

The `Factory` creates standardized responses:

```go
factory := dr.New(
    dr.WithLogger(logger),
    dr.WithFormatter(formatter.NewJSON()),
    dr.WithVerbosity(true),
)

// Success responses
resp := factory.Success(ctx, data) // 200 OK
resp := factory.Created(ctx, data, location) // 201 Created
resp := factory.Accepted(ctx, data) // 202 Accepted
resp := factory.NoContent(ctx) // 204 No Content

// Error responses
resp := factory.BadRequest(ctx, message) // 400
resp := factory.Unauthorized(ctx, message) // 401
resp := factory.Forbidden(ctx, message) // 403
resp := factory.NotFound(ctx, message) // 404
resp := factory.Conflict(ctx, message) // 409
resp := factory.InternalError(ctx, err) // 500

// Validation errors
resp := factory.ValidationError(
	    ctx, 
	    message, 
	    map[string][]string{"email": {"required", "invalid format"}}, 
	)
```

### Immutable Response Pattern

All response modifications return a new instance:

```go
resp := factory.Success(ctx, data).
WithHeader("X-Request-ID", requestID).
WithSecurityHeaders().
WithCORS("*", "GET,POST", "Content-Type").
WithCacheControl("no-cache")
```

### Middleware System

Create composable middleware chains:

```go
// Custom middleware
func AuthMiddleware(next dr.Handler) dr.Handler {
return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
    token := r.Header.Get("Authorization")
    if token == "" {
        return f.Unauthorized(r.Context(), "missing token")
    }
    // Validate token...
    
    return next.Handle(r, f)
})

// Use middleware
r.WithMiddleware(
    middleware.RequestTimer(),
    middleware.Recovery(),
    AuthMiddleware,
    middleware.LoggingDefault(),
)
```

### Chi Middleware Integration

Wrap existing chi middleware:

```go
import chimiddleware "github.com/go-chi/chi/v5/middleware"

r.WithMiddleware(
    dr.WrapMiddleware(chimiddleware.RequestID),
    dr.WrapMiddleware(chimiddleware.RealIP),
    dr.WrapMiddleware(chimiddleware.Compress(5)),
)
```

## Advanced Usage

### Custom Logging Template

```go
loggingMW, err := middleware.Logging(&middleware.LoggingConfig{
    Template: {{.Method}} {{.URI}} {{.Status}} {{.Duration}} - {{.Custom.UserID}},
    ContextFields: map[string]middleware.ContextValueFunc{
        "UserID": func(ctx context.Context) interface{} {
            if user := ctx.Value("user"); user != nil {
                return user.(*User).ID
            }
			
            return "-"
        },
    },
})
```
### Metrics Collection

```go
type MyMetricsService struct{}

func (m *MyMetricsService) Responded(data middleware.MetricsData) {
    // Send to Prometheus, StatsD, etc.
    prometheusCounter.WithLabelValues(
        data.Method,
        data.Route,
        strconv.Itoa(data.StatusCode),
    ).Inc()
}

r.WithMiddleware(
    middleware.RequestTimer(),
    middleware.Measurement(metricsService, chi.RouteContext),
)
```

### Custom Error Builder

```go
customErrorBuilder := func(ctx context.Context, status int, message string, details any) any {
    return map[string]interface{}{
        "error": map[string]interface{}{
        "code": status,
        "message": message,
        "details": details,
        "trace_id": ctx.Value("trace_id"),
        },
    }
}

factory := dr.New(dr.WithErrorBuilder(customErrorBuilder))
```

### Binary File Responses

```go
func downloadFile(r *http.Request, f *dr.Factory) dr.DataResponse {
    return f.File(r.Context(), "/path/to/file.pdf")
}

func streamData(r *http.Request, f *dr.Factory) dr.DataResponse {
    reader := getDataReader()
	
    return f.Binary(r.Context(), reader, "data.csv", size)
}
```

### Request Context Values

```go
// Store request start time
r.WithMiddleware(middleware.RequestTimer())

// Access in handlers or middleware
start := middleware.GetRequestStartTime(r.Context())
duration := time.Since(start)
```

## Middleware

### Built-in Middleware

| Middleware | Description |
|------------|-------------|
| `RequestTimer()` | Captures request start time in context |
| `Recovery()` | Recovers from panics and returns 500 |
| `Logging(cfg)` | Customizable access log with templates |
| `LoggingDefault()` | Default Apache-style access log |
| `Measurement(service, pattern)` | Metrics collection |

### Creating Custom Middleware

```go
func RateLimitMiddleware(limiter *rate.Limiter) dr.Middleware {
    return func(next dr.Handler) dr.Handler {
        return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
            if !limiter.Allow() {
                return f.Error(r.Context(), http.StatusTooManyRequests, "rate limit exceeded")
            }
            
            return next.Handle(r, f)
        })
    }
}
```

## API Documentation

### Factory Methods

| Method | HTTP Status | Description |
|--------|-------------|-------------|
| `Success(ctx, data)` | 200 | Success response |
| `Created(ctx, data, location)` | 201 | Resource created |
| `Accepted(ctx, data)` | 202 | Request accepted |
| `NoContent(ctx)` | 204 | No content |
| `BadRequest(ctx, msg)` | 400 | Bad request |
| `Unauthorized(ctx, msg)` | 401 | Unauthorized |
| `Forbidden(ctx, msg)` | 403 | Forbidden |
| `NotFound(ctx, msg)` | 404 | Not found |
| `Conflict(ctx, msg)` | 409 | Conflict |
| `ValidationError(ctx, msg, errors)` | 422 | Validation error |
| `InternalError(ctx, err)` | 500 | Internal error |
| `ServiceUnavailable(ctx, msg)` | 503 | Service unavailable |

### Response Methods

| Method | Description |
|--------|-------------|
| `WithStatusCode(code)` | Set HTTP status code |
| `WithHeader(key, value)` | Add header |
| `WithHeaders(headers)` | Add multiple headers |
| `WithContentType(ct)` | Set content type |
| `WithCORS(origin, methods, headers)` | Add CORS headers |
| `WithSecurityHeaders()` | Add security headers |
| `WithCacheControl(value)` | Set cache control |
| `WithData(data)` | Replace response data |

## Examples

See the [examples](example/) directory for complete working examples:

- [Basic Usage](examples/basic/)
- [Chi Integration](examples/chi/)
- [Custom Middleware](middleware/)
- [Custom Handlers](handler/)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

