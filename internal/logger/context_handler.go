package logger

import (
	"context"
	"log/slog"
	"sync"
)

type ctxKey string

const (
	logFields ctxKey = "log_fields"
)

var (
	once sync.Once
	l    *slog.Logger
)

type ContextHandler struct {
	slog.Handler
}

func InitializeLogger(logger *slog.Logger) {
	once.Do(func() {
		slog.SetDefault(logger)
	})
}

func Error(err error, msg string, args ...any) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	argsWithErrorArg := append([]any{"err", errMsg}, args...)
	slog.Error(msg, argsWithErrorArg...)
}

func ErrorCtx(ctx context.Context, err error, msg string, args ...any) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	argsWithErrorArg := append([]any{"err", errMsg}, args...)
	slog.ErrorContext(ctx, msg, argsWithErrorArg...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

// Handle adds contextual information based on context before calling underlying handler
func (c *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(logFields).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}

	return c.Handler.Handle(ctx, r)
}

// func AddToLogContext(parent context.Context, attr slog.Attr) context.Context {
func AddToLogContext(parent context.Context, key string, value any) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(logFields).([]slog.Attr); ok {
		v := v
		// iterate over the attributes and update the value if the key exists
		for idx := 0; idx < len(v); idx++ {
			if v[idx].Key == key {
				v[idx].Value = slog.AnyValue(value)
				return context.WithValue(parent, logFields, v)
			}
		}
		v = append(v, slog.Attr{
			Key:   key,
			Value: slog.AnyValue(value),
		})
		return context.WithValue(parent, logFields, v)
	}

	v := []slog.Attr{{
		Key:   key,
		Value: slog.AnyValue(value),
	}}

	return context.WithValue(parent, logFields, v)
}
