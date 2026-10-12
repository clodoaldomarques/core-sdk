package otel

import (
	"context"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/meter"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	OtlpUrl string
)

func init() {
	OtlpUrl = env.GetString(env.OTEL_EXPORTER_ENDPOINT, "")
}

func Start(ctx context.Context) {
	tp := tracer.InitTracer(ctx, OtlpUrl)
	otel.SetTracerProvider(tp)
	mp := meter.InitMeter(ctx, OtlpUrl)
	otel.SetMeterProvider(mp)

	if err := runtime.Start(
		runtime.WithMeterProvider(mp),
		runtime.WithMinimumReadMemStatsInterval(15*time.Second),
	); err != nil {
		logger.Error(ctx, "Erro ao iniciar runtime metrics", logger.Fields{
			"error": err.Error(),
		})
	}

}

func Shutdown(ctx context.Context) error {
	if tp := otel.GetTracerProvider().(*trace.TracerProvider); tp != nil {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(ctx, "Erro no shutdown do TracerProvider", logger.Fields{
				"error": err.Error(),
			})
			return err
		}
	}

	if mp := otel.GetMeterProvider().(*metric.MeterProvider); mp != nil {
		if err := mp.Shutdown(ctx); err != nil {
			logger.Error(ctx, "Erro no shutdown do MeterProvider", logger.Fields{
				"error": err.Error(),
			})
			return err
		}
	}
	return nil
}
