package otel

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/clodoaldomarques/core-sdk/pkg/opentelemetry/logger"
	"github.com/clodoaldomarques/core-sdk/pkg/opentelemetry/meter"
	"github.com/clodoaldomarques/core-sdk/pkg/opentelemetry/tracer"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	OtlpUrl string
	tp      *trace.TracerProvider
	mp      *metric.MeterProvider
	lp      *log.LoggerProvider
	lg      logr.Logger
)

func init() {
	OtlpUrl = env.GetString(env.OTEL_EXPORTER_ENDPOINT, "")
}

func Start(ctx context.Context) {
	tp = tracer.InitTracer(ctx, OtlpUrl)
	mp = meter.InitMeter(ctx, OtlpUrl)
	lp, lg = logger.InitLogger(ctx, OtlpUrl)
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetLogger(lg)
}

func Shutdown(ctx context.Context) error {
	if err := tp.Shutdown(ctx); err != nil {
		logger.Error(ctx, "Erro no shutdown do TracerProvider", logger.Fields{
			"error": err.Error(),
		})
		return err
	}

	if err := mp.Shutdown(ctx); err != nil {
		logger.Error(ctx, "Erro no shutdown do MeterProvider", logger.Fields{
			"error": err.Error(),
		})
		return err
	}

	if err := lp.Shutdown(ctx); err != nil {
		logger.Error(ctx, "Erro no shutdown do LoggerProvider", logger.Fields{
			"error": err.Error(),
		})
		return err
	}

	return nil
}
