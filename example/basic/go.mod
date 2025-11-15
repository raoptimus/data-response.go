module github.com/raoptimus/data-response.go/example/basic

go 1.25.4

replace (
	github.com/raoptimus/data-response.go/pkg/logger v0.0.0 => ../../pkg/logger
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog v0.0.0 => ../../pkg/logger/adapter/slog
	github.com/raoptimus/data-response.go/v2 v2.0.0 => ../../
)

require (
	github.com/raoptimus/data-response.go v1.4.2
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog v0.0.0
	github.com/raoptimus/data-response.go/v2 v2.0.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/raoptimus/data-response.go/pkg/logger v0.0.0 // indirect
)
