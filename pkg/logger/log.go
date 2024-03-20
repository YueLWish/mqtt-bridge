package logger

import (
	"context"
	"github.com/yuelwish/mqtt-bridge/pkg/setting"
	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(conf *setting.AppConfig) (logger *Logger, syncFn func(), err error) {
	zLogger, err := NewZapLogger(conf.Log)
	if err != nil {
		return nil, nil, err
	}

	zap.ReplaceGlobals(zLogger)

	return &Logger{Logger: zLogger}, func() { _ = logger.Sync() }, nil
}

type logKey struct{}

// NewContext 创建一个存放 log 的上下文
func (l *Logger) NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &Logger{l.WithContext(ctx).With(fields...)})
}

func (l *Logger) NewContextWithLogger(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &Logger{Logger: l.WithContext(ctx).With(fields...)})
}

// WithContext 从上下中获取 zap 的日志实例
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(logKey{}).(*Logger); ok {
		return logger
	} else {
		return l
	}
}

func DefaultLogger() *Logger {
	return &Logger{Logger: zap.L()}
}

func NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &Logger{Logger: zap.L().With(fields...)})
}

// NewContextWithLogger 新增上下文，在原有Logger的基础上 新增 Field
func NewContextWithLogger(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey{}, &Logger{Logger: WithContext(ctx).With(fields...)})
}

func WithContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(logKey{}).(*Logger); ok {
		return logger
	} else {
		return &Logger{Logger: zap.L()}
	}
}
