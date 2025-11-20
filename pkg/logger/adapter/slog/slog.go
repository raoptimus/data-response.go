package slog

import (
	"context"
	"log/slog"
)

// Slog адаптирует log/slog к интерфейсу Logger
type Slog struct {
	logger *slog.Logger
}

// New создает новый adapter для slog
func New(logger *slog.Logger) *Slog {
	if logger == nil {
		logger = slog.Default()
	}
	return &Slog{logger: logger}
}

// Debug логирует debug-level сообщение
func (a *Slog) Debug(ctx context.Context, msg string, args ...any) {
	a.logger.DebugContext(ctx, msg, args...)
}

// Info логирует info-level сообщение
func (a *Slog) Info(ctx context.Context, msg string, args ...any) {
	a.logger.InfoContext(ctx, msg, args...)
}

// Warn логирует warning-level сообщение
func (a *Slog) Warn(ctx context.Context, msg string, args ...any) {
	a.logger.WarnContext(ctx, msg, args...)
}

// Error логирует error-level сообщение
func (a *Slog) Error(ctx context.Context, msg string, args ...any) {
	a.logger.ErrorContext(ctx, msg, args...)
}

// WithAttrs возвращает новый adapter с дополнительными атрибутами
func (a *Slog) WithAttrs(attrs ...slog.Attr) *Slog {
	return &Slog{
		logger: a.logger.With(attrsToAny(attrs)...),
	}
}

// WithGroup возвращает новый adapter с группой
func (a *Slog) WithGroup(name string) *Slog {
	return &Slog{
		logger: a.logger.WithGroup(name),
	}
}

// Helper для конвертации slog.Attr в any
func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}
