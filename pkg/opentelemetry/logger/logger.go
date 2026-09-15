package logger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
)

var (
	once       sync.Once
	slogLogger *slog.Logger
	OtlService string
)

func init() {
	OtlService = env.GetString(env.OTEL_SERVICE_NAME, "")
}

func InitLogger(ctx context.Context, otlpUrl string) (*log.LoggerProvider, logr.Logger) {
	exporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(otlpUrl),
	)
	if err != nil {
		panic(err)
	}

	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(log.WithProcessor(processor))

	once.Do(func() {
		slogLogger = otelslog.NewLogger(fmt.Sprintf("%s/logs", OtlService), otelslog.WithLoggerProvider(provider))
		slog.SetDefault(slogLogger)
	})

	logrLogger := logr.FromSlogHandler(slogLogger.Handler())

	return provider, logrLogger
}

type Fields map[string]any

func Info(ctx context.Context, message string, fields Fields) {
	slogLogger.InfoContext(ctx, message, toAttrs(fields)...)
}

func Warn(ctx context.Context, message string, fields Fields) {
	slogLogger.WarnContext(ctx, message, toAttrs(fields)...)
}

func Error(ctx context.Context, message string, fields Fields) {
	slogLogger.ErrorContext(ctx, message, toAttrs(fields)...)
}

func Fatal(ctx context.Context, message string, fields Fields) {
	slogLogger.ErrorContext(ctx, message, toAttrs(fields)...)
	// slog não tem Fatal nativo; se quiser, pode chamar os.Exit(1) aqui.
}

func toAttrs(fields Fields) []any {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return attrs
}
