package log

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
)

func Init() {
}

var defaultLogger atomic.Value

// Default returns the default Logger.
func Default() *Logger {
	log, ok := defaultLogger.Load().(*Logger)
	if ok {
		return log
	}
	log = newDefault()
	defaultLogger.Store(log)
	return log
}

// Close close the defalut Logger
func Close() {
	d := Default()
	if d != nil {
		d.Close()
	}
}
func SetDefaultLogger(l *Logger) {
	defaultLogger.Store(l)
}
func SetDefaultLoggerLevel(level slog.Level) {
	Default().SetLevel(level)
}

// byte uint
func SetDefaultLoggerMaxFileSize(max int) {
	Default().SetMaxFileSize(max)
}
func AddExtractAttrsFromContextFn(fn ExtractAttrFn) {
	Default().AddExtractAttrsFromContextFn(fn)
}

func newDefault() *Logger {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	lv := slog.LevelInfo
	switch os.Getenv("GO_LOG") {
	case "error":
		lv = slog.LevelError
	case "warn":
		lv = slog.LevelWarn
	case "debug":
		lv = slog.LevelDebug
	}

	if os.Getenv("GO_LOG_OUT") == "console" {
		root = ""
	}
	logger, err := New(root, lv)

	if err != nil {
		panic(err)
	}
	return logger
}

func Debug(args ...any) {
	Default().log(context.Background(), slog.LevelDebug, args...)
}

func Info(args ...any) {
	Default().log(context.Background(), slog.LevelInfo, args...)
}
func Warn(args ...any) {
	Default().log(context.Background(), slog.LevelWarn, args...)
}

func Error(args ...any) {
	Default().log(context.Background(), slog.LevelError, args...)
}

func DebugContext(ctx context.Context, args ...any) {
	Default().log(ctx, slog.LevelDebug, args...)
}

func InfoContext(ctx context.Context, args ...any) {
	Default().log(ctx, slog.LevelInfo, args...)
}

func WarnContext(ctx context.Context, args ...any) {
	Default().log(ctx, slog.LevelWarn, args...)
}

func ErrorContext(ctx context.Context, args ...any) {
	Default().log(ctx, slog.LevelError, args...)
}

func Debugf(format string, args ...any) {
	Default().logf(context.Background(), slog.LevelDebug, format, args...)
}

func Infof(format string, args ...any) {
	Default().logf(context.Background(), slog.LevelInfo, format, args...)
}
func Warnf(format string, args ...any) {
	Default().logf(context.Background(), slog.LevelWarn, format, args...)
}

func Errorf(format string, args ...any) {
	Default().logf(context.Background(), slog.LevelError, format, args...)
}

func DebugfContext(ctx context.Context, format string, args ...any) {
	Default().logf(ctx, slog.LevelDebug, format, args...)
}

func InfofContext(ctx context.Context, format string, args ...any) {
	Default().logf(ctx, slog.LevelInfo, format, args...)
}

func WarnfContext(ctx context.Context, format string, args ...any) {
	Default().logf(ctx, slog.LevelWarn, format, args...)
}

func ErrorfContext(ctx context.Context, format string, args ...any) {
	Default().logf(ctx, slog.LevelError, format, args...)
}

