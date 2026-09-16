package otel

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/logger"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/meter"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
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
	otel.SetTracerProvider(tracer.InitTracer(ctx, OtlpUrl))
	otel.SetMeterProvider(meter.InitMeter(ctx, OtlpUrl))
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
