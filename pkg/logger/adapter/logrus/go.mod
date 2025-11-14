module github.com/raoptimus/data-response.go/pkg/logger/adapter/logrus

go 1.25

require (
	github.com/raoptimus/data-response.go/pkg/logger v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.38.0 // indirect

replace github.com/raoptimus/data-response.go/pkg/logger v0.0.0 => ../../
