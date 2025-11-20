package dataresponse

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raoptimus/data-response.go/v2/formatter"
	"github.com/raoptimus/data-response.go/v2/response"
)

// Benchmark Factory methods
func BenchmarkFactory_Success(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	ctx := context.Background()
	data := map[string]string{"key": "value"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory.Success(ctx, data)
	}
}

func BenchmarkFactory_Error(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory.Error(ctx, http.StatusBadRequest, "error message")
	}
}

func BenchmarkFactory_InternalError(b *testing.B) {
	factory := New(
		WithFormatter(defaultFormatter()),
		WithVerbosity(false), // Disable logging for benchmark
	)
	ctx := context.Background()
	err := response.NewError(http.StatusInternalServerError, "test error")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory.InternalError(ctx, err)
	}
}

func BenchmarkFactory_ValidationError(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	ctx := context.Background()
	errors := map[string][]string{
		"email":    {"required", "invalid"},
		"password": {"too short"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory.ValidationError(ctx, "validation failed", errors)
	}
}

// Benchmark Response operations
func BenchmarkDataResponse_WithHeader(b *testing.B) {
	factory := New()
	resp := factory.Success(context.Background(), nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = resp.WithHeader("X-Custom-Header", "value")
	}
}

func BenchmarkDataResponse_WithHeaders(b *testing.B) {
	factory := New()
	resp := factory.Success(context.Background(), nil)
	headers := http.Header{
		"X-Header-1": []string{"value1"},
		"X-Header-2": []string{"value2"},
		"X-Header-3": []string{"value3"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = resp.WithHeaders(headers)
	}
}

func BenchmarkDataResponse_WithSecurityHeaders(b *testing.B) {
	factory := New()
	resp := factory.Success(context.Background(), nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = resp.WithSecurityHeaders()
	}
}

func BenchmarkDataResponse_WithCORS(b *testing.B) {
	factory := New()
	resp := factory.Success(context.Background(), nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = resp.WithCORS("*", "GET,POST,PUT", "Content-Type,Authorization")
	}
}

func BenchmarkHandler_NoContent(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.NoContent(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = handler.Handle(req, factory)
	}
}

// Benchmark Handler operations
func BenchmarkHandler_Simple(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"result": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = handler.Handle(req, factory)
	}
}

func BenchmarkHandler_WithMiddleware(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))

	mw := func(next Handler) Handler {
		return HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
			resp := next.Handle(r, f)
			return resp.WithHeader("X-Middleware", "applied")
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"result": "ok"})
	})

	finalHandler := mw(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = finalHandler.Handle(req, factory)
	}
}

func BenchmarkHandler_ChainedMiddleware(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))

	mw1 := func(next Handler) Handler {
		return HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
			return next.Handle(r, f).WithHeader("X-MW-1", "1")
		})
	}

	mw2 := func(next Handler) Handler {
		return HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
			return next.Handle(r, f).WithHeader("X-MW-2", "2")
		})
	}

	mw3 := func(next Handler) Handler {
		return HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
			return next.Handle(r, f).WithHeader("X-MW-3", "3")
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"result": "ok"})
	})

	chain := Chain(handler, mw1, mw2, mw3)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = chain.Handle(req, factory)
	}
}

// Benchmark WrapHandler (full HTTP request cycle)
func BenchmarkWrapHandler_FullCycle(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"result": "ok"})
	})

	httpHandler := WrapHandler(handler, factory)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		httpHandler.ServeHTTP(w, req)
	}
}

// Benchmark WrapMiddleware (chi integration)
func BenchmarkWrapMiddleware(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))

	chiMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Chi-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	}

	wrappedMW := WrapMiddleware(chiMW)
	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"result": "ok"})
	})

	finalHandler := wrappedMW(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = finalHandler.Handle(req, factory)
	}
}

// Benchmark Error creation
func BenchmarkNewError(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = response.NewError(http.StatusBadRequest, "error message")
	}
}

func BenchmarkWrapError(b *testing.B) {
	originalErr := response.NewError(http.StatusInternalServerError, "original")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = response.WrapError(http.StatusInternalServerError, originalErr, "wrapped")
	}
}

// Benchmark formatter operations
func BenchmarkFormatter_JSON(b *testing.B) {
	f := formatter.NewJSON()
	factory := New(WithFormatter(f))
	resp := factory.Success(context.Background(), map[string]interface{}{
		"id":      123,
		"name":    "John Doe",
		"email":   "john@example.com",
		"active":  true,
		"balance": 99.99,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		formatted, err := f.Format(resp)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, formatted.Stream)
	}
}

// Benchmark different data sizes
func BenchmarkFactory_Success_SmallData(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	ctx := context.Background()
	data := map[string]string{"key": "value"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory.Success(ctx, data)
	}
}

func BenchmarkFactory_Success_MediumData(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	ctx := context.Background()
	data := map[string]interface{}{
		"id":         123,
		"name":       "John Doe",
		"email":      "john@example.com",
		"address":    "123 Main St",
		"city":       "New York",
		"country":    "USA",
		"age":        30,
		"active":     true,
		"balance":    1000.50,
		"created_at": "2025-11-19T12:00:00Z",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory.Success(ctx, data)
	}
}

func BenchmarkFactory_Success_LargeData(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	ctx := context.Background()

	// Create array of 100 items
	items := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"id":     i,
			"name":   "Item " + string(rune(i)),
			"value":  i * 10,
			"active": i%2 == 0,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory.Success(ctx, items)
	}
}

// Benchmark parallel execution
func BenchmarkFactory_Success_Parallel(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	data := map[string]string{"key": "value"}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			_ = factory.Success(ctx, data)
		}
	})
}

func BenchmarkHandler_FullCycle_Parallel(b *testing.B) {
	factory := New(WithFormatter(defaultFormatter()))
	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"result": "ok"})
	})

	httpHandler := WrapHandler(handler, factory)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			httpHandler.ServeHTTP(w, req)
		}
	})
}
