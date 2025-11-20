package dataresponse

import "context"

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (l NoOpLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (l NoOpLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (l NoOpLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (l NoOpLogger) Error(ctx context.Context, msg string, args ...any) {}
