package zap

import (
	"context"

	"go.uber.org/zap"
)

// Zap adapts zap.SugaredLogger to dataresponse.Logger interface.
type Zap struct {
	logger *zap.SugaredLogger
}

// NewZap creates a new zap adapter from SugaredLogger.
func NewZap(logger *zap.SugaredLogger) *Zap {
	if logger == nil {
		l, _ := zap.NewProduction()
		logger = l.Sugar()
	}
	return &Zap{logger: logger}
}

// NewZapFromLogger creates adapter from regular zap.Logger.
func NewZapFromLogger(logger *zap.Logger) *Zap {
	return &Zap{logger: logger.Sugar()}
}

// Debug logs a debug-level message.
func (a *Zap) Debug(_ context.Context, msg string, args ...any) {
	a.logger.Debugw(msg, args...)
}

// Info logs an info-level message.
func (a *Zap) Info(_ context.Context, msg string, args ...any) {
	a.logger.Infow(msg, args...)
}

// Warn logs a warning-level message.
func (a *Zap) Warn(_ context.Context, msg string, args ...any) {
	a.logger.Warnw(msg, args...)
}

// Error logs an error-level message.
func (a *Zap) Error(_ context.Context, msg string, args ...any) {
	a.logger.Errorw(msg, args...)
}
