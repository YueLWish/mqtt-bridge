package logger

import (
	"github.com/yuelwish/mqtt-bridge/pkg/setting"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func NewZapLogger(conf setting.Log) (logger *zap.Logger, err error) {
	var multiCore = make([]zapcore.Core, 0, 2)

	var l = new(zapcore.Level) // 日志级别
	if err = l.UnmarshalText([]byte(conf.Level)); err != nil {
		return
	}
	encoder := getEncoder() // 编码器

	{ // 日志写入 同步任务
		writeSyncer := getFileWriteSyncer(conf.FileName, conf.MaxSize, conf.MaxAge, conf.MaxBackups)
		multiCore = append(multiCore, zapcore.NewCore(encoder, writeSyncer, l))
	}

	if conf.Console { // 控制台 输出
		multiCore = append(multiCore, zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), l))
	}

	// 创建日志
	logger = zap.New(
		zapcore.NewTee(multiCore...), // 创建 core
		zap.AddCaller(),              // 在日志记录中增加调用行号和函数名
		//zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel), // 错误打印堆栈
	)
	return logger, nil
}

func getEncoder() zapcore.Encoder {
	conf := zap.NewProductionEncoderConfig()
	conf.EncodeTime = zapcore.RFC3339TimeEncoder        // 更改时间编码
	conf.EncodeLevel = zapcore.CapitalColorLevelEncoder // 在日志文件中使用大写字母记录日志级别 CapitalLevelEncoder
	conf.EncodeDuration = zapcore.SecondsDurationEncoder
	conf.EncodeCaller = zapcore.ShortCallerEncoder

	// 按行的日志格式  ==> 2022-07-15T15:14:17.064985963+08:00     INFO    carbon/main.go.go:69       Start Server    {"address": ":8001"}
	return zapcore.NewConsoleEncoder(conf)
}

func getFileWriteSyncer(fileName string, maxSize int, maxAge int, maxBackups int) zapcore.WriteSyncer {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fileName,   // 文件路径
		MaxSize:    maxSize,    // 单个文件最大尺寸，默认单位 M
		MaxAge:     maxAge,     // 最大时间，默认单位 day
		MaxBackups: maxBackups, // 最多保留备份数量
		LocalTime:  true,       // 使用本地时间
		Compress:   true,       // 是否压缩
	}
	return zapcore.AddSync(lumberjackLogger)
}
