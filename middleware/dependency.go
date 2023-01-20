package middleware

import (
	"context"
)

type Logger interface {
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)

	Error(v ...any)
	Errorf(format string, v ...any)
	Errorln(v ...any)
}

type LoggerWithContext interface {
	Logger
	WithContext(ctx context.Context) Logger
}
