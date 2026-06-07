package opentelemetry

import (
	"context"
	"sync"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	once    sync.Once
	OtlpUrl string
)

func init() {
	once.Do(func() {
		OtlpUrl = env.GetString(env.OTEL_EXPORTER_ENDPOINT, "")
	})
}

func InitTracer(ctx context.Context) *trace.TracerProvider {
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(OtlpUrl),
		otlptracegrpc.WithTimeout(2*time.Second),
	)
	if err != nil {
		panic(err)
	}
	provide := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(traceExporter),
	)

	otel.SetTracerProvider(provide)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return provide
}

func InitMeter(ctx context.Context) *metric.MeterProvider {
	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(OtlpUrl),
	)

	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithTimeout(2*time.Second))),
	)
	return meterProvider
}

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

func Shutdown(ctx context.Context) error {
	tp := otel.GetTracerProvider()
	if tp, ok := tp.(*trace.TracerProvider); ok {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(ctx, "Erro no shutdown do TracerProvider", logger.Fields{
				"error": err.Error(),
			})
			return err
		}
	}

	mp := otel.GetMeterProvider()
	if mp, ok := mp.(*metric.MeterProvider); ok {
		if err := mp.Shutdown(ctx); err != nil {
			logger.Error(ctx, "Erro no shutdown do MeterProvider", logger.Fields{
				"error": err.Error(),
			})
			return err
		}
	}

	return nil
}
