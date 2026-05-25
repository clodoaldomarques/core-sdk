package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once sync.Once
)

type Fields map[string]any

func init() {
	once.Do(func() {
		singleton, _ := NewZapLogger(zap.AddCallerSkip(1))
		zap.ReplaceGlobals(singleton)
	})
}

func NewZapLogger(options ...zap.Option) (*zap.Logger, error) {
	return loggerConfig().Build(options...)
}

func Info(ctx context.Context, message string, fields Fields) {
	zap.L().Info(message, buildFields(fields)...)
}
func Warn(ctx context.Context, message string, fields Fields) {
	zap.L().Warn(message, buildFields(fields)...)
}
func Error(ctx context.Context, message string, fields Fields) {
	zap.L().Error(message, buildFields(fields)...)
}
func Fatal(ctx context.Context, message string, fields Fields) {
	zap.L().Fatal(message, buildFields(fields)...)
}

func buildFields(fields Fields) []zapcore.Field {
	return []zapcore.Field{
		zap.Any("Attributes", fields),
	}
}

func loggerConfig() zap.Config {
	al := zap.NewAtomicLevel()
	return zap.Config{
		Level:       al,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    loggerEncoder(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func loggerEncoder() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "Message",
		LevelKey:       "Level",
		TimeKey:        "Timestamp",
		NameKey:        "Logger",
		CallerKey:      "Caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "Stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
