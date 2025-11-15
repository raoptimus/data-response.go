module github.com/raoptimus/data-response.go/v2

go 1.25.4

require (
	github.com/andybalholm/brotli v1.2.0
	github.com/json-iterator/go v1.1.12
	github.com/pkg/errors v0.9.1
	github.com/raoptimus/data-response.go/pkg/logger v0.0.0
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog v0.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/raoptimus/data-response.go/pkg/logger v0.0.0 => ./pkg/logger
	github.com/raoptimus/data-response.go/pkg/logger/adapter/slog v0.0.0 => ./pkg/logger/adapter/slog
)
