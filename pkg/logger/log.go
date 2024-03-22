package logger

import (
	"context"
	"github.com/yuelwish/mqtt-bridge/pkg/setting"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	With(fields ...zap.Field) *zap.Logger
	WithOptions(opts ...zap.Option) *zap.Logger
	Log(lvl zapcore.Level, msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
}

type xlog struct {
	*zap.Logger
}

func NewLogger(conf *setting.AppConfig) (_xlog Logger, syncFn func(), err error) {
	zLogger, err := NewZapLogger(conf.Log)
	if err != nil {
		return nil, nil, err
	}

	zap.ReplaceGlobals(zLogger)
	return &xlog{Logger: zLogger}, func() { _ = _xlog.Sync() }, nil
}

type logKey struct{}

// NewContext 创建一个存放 log 的上下文
func (l *xlog) NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &xlog{l.WithContext(ctx).With(fields...)})
}

func (l *xlog) NewContextWithLogger(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &xlog{Logger: l.WithContext(ctx).With(fields...)})
}

// WithContext 从上下中获取 zap 的日志实例
func (l *xlog) WithContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(logKey{}).(Logger); ok {
		return logger
	} else {
		return l
	}
}

func DefaultLogger(opts ...zap.Option) Logger {
	return &xlog{Logger: zap.L().WithOptions(opts...)}
}

func NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &xlog{Logger: zap.L().With(fields...)})
}

// NewContextWithLogger 新增上下文，在原有Logger的基础上 新增 Field
func NewContextWithLogger(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &xlog{Logger: WithContext(ctx).With(fields...)})
}

func WithContext(ctx context.Context) Logger {
	if log, ok := ctx.Value(logKey{}).(Logger); ok {
		return log
	} else {
		return &xlog{Logger: zap.L()}
	}
}

func With(fields ...zap.Field) Logger {
	return DefaultLogger().With(fields...)
}
func Log(lvl zapcore.Level, msg string, fields ...zap.Field) {
	DefaultLogger(zap.AddCallerSkip(1)).Log(lvl, msg, fields...)
}
func Debug(msg string, fields ...zap.Field) {
	DefaultLogger(zap.AddCallerSkip(1)).Debug(msg, fields...)
}
func Info(msg string, fields ...zap.Field) {
	DefaultLogger(zap.AddCallerSkip(1)).Info(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	DefaultLogger(zap.AddCallerSkip(1)).Warn(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	DefaultLogger(zap.AddCallerSkip(1)).Error(msg, fields...)
}
func Fatal(msg string, fields ...zap.Field) {
	DefaultLogger(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}
