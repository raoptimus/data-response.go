module github.com/raoptimus/data-response.go/pkg/chiadapter

go 1.25.4

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/raoptimus/data-response.go/v2 v2.0.0
)

require (
	github.com/pkg/errors v0.9.1 // indirect
	github.com/raoptimus/data-response.go/pkg/logger v0.0.0 // indirect
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog v0.0.0 // indirect
)

replace (
	github.com/raoptimus/data-response.go/pkg/logger v0.0.0 => ../logger
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog v0.0.0 => ../logger/adapter/slog
	github.com/raoptimus/data-response.go/v2 v2.0.0 => ../../
)
