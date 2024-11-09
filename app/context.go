package app

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

var ctxKeyLogger = ctxKey{}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	v := ctx.Value(ctxKeyLogger)
	if v == nil {
		return nil
	}
	logger, ok := v.(*zap.Logger)
	if ok {
		return logger
	}
	return nil
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, logger)
}
