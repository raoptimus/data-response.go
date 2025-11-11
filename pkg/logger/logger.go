package logger

import "context"

// Logger defines the logging interface.
// Compatible with slog, logrus, zap, and zerolog through adapters.
type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (l NoOpLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (l NoOpLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (l NoOpLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (l NoOpLogger) Error(ctx context.Context, msg string, args ...any) {}
