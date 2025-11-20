module github.com/raoptimus/data-response.go/example/chi

go 1.25.4

replace (
	github.com/raoptimus/data-response.go/pkg/chiadapter => ../../pkg/chiadapter
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog => ../../pkg/logger/adapter/slog
	github.com/raoptimus/data-response.go/v2 => ../../
)

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/json-iterator/go v1.1.12
	github.com/raoptimus/data-response.go/pkg/chiadapter v0.0.0-00010101000000-000000000000
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog v0.0.0-00010101000000-000000000000
	github.com/raoptimus/data-response.go/v2 v2.0.0-00010101000000-000000000000
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
)
