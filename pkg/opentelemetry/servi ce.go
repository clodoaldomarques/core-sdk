package opentelemetry

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"go.opentelemetry.io/otel"
)

func Start(ctx context.Context) {
	tracerProvicer := InitTracer(ctx)
	defer func() {
		if err := tracerProvicer.Shutdown(ctx); err != nil {
			logger.Error(ctx, "error shutdown trace provider", logger.Fields{"error": err.Error()})
		}
	}()

	meterProvicer := InitMeter(ctx)
	defer func() {
		if err := meterProvicer.Shutdown(ctx); err != nil {
			logger.Error(ctx, "error shutdonw meter provider", logger.Fields{"error": err.Error()})
		}
	}()

	otel.SetTracerProvider(tracerProvicer)
	otel.SetMeterProvider(meterProvicer)
}
