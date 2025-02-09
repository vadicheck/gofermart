package logger

import (
	"context"
	"github.com/vadicheck/gofermart/pkg/logger/formatter"
	"github.com/vadicheck/gofermart/pkg/logger/options"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

func (l *commonLogger) InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Info(msg)
}

func (l *commonLogger) TraceCtx(ctx context.Context, msg string, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Trace(msg)
}

func (l *commonLogger) WarnCtx(ctx context.Context, msg string, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Warn(msg)
}

func (l *commonLogger) DebugCtx(ctx context.Context, msg string, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Debug(msg)
}

func (l *commonLogger) ErrorMessageCtx(ctx context.Context, msg string, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Error(msg)
}

func (l *commonLogger) ErrorCtx(ctx context.Context, err error, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Errorf("[ERR] %s stacktrace: %s", err, string(debug.Stack()))
}

func (l *commonLogger) FatalCtx(ctx context.Context, err error, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Fatalf("[ERR] %s stacktrace: %s", err, string(debug.Stack()))
}

func (l *commonLogger) PanicCtx(ctx context.Context, err error, fields ...interface{}) {
	l.setFieldsWithContext(ctx, options.WithExtras(fields)).Panicf("[ERR] %s stacktrace: %s", err, string(debug.Stack()))
}

func (l *commonLogger) SetOptionsToCtx(ctx context.Context, optValues ...options.LoggerOption) context.Context {
	opts := &options.LoggerOptions{}
	optsFromCtx := l.OptionsFromCtx(ctx)
	if optsFromCtx != nil {
		opts = optsFromCtx
	}
	for _, setOption := range optValues {
		setOption(opts)
	}
	return context.WithValue(ctx, contextKey(LogOptionsContextKey), *opts)
}

func (l *commonLogger) OptionsFromCtx(ctx context.Context) *options.LoggerOptions {
	if ctx == nil {
		return nil
	}
	opts, ok := ctx.Value(contextKey(LogOptionsContextKey)).(options.LoggerOptions)
	if ok {
		return &opts
	}
	return nil
}

func (l *commonLogger) setFieldsWithContext(ctx context.Context, optValues ...options.LoggerOption) *logrus.Entry {
	opts := &options.LoggerOptions{}
	optsFromCtx := l.OptionsFromCtx(ctx)
	if optsFromCtx != nil {
		opts = optsFromCtx
	}

	for _, setOption := range optValues {
		setOption(opts)
	}
	return l.Console.WithField(formatter.LogOptionsField, *opts)
}
