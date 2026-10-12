package meter

import (
	"context"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/env"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitMeter(ctx context.Context, otlpUrl string) *metric.MeterProvider {
	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(otlpUrl),
	)

	res, err := resource.New(ctx,
		// Detecção automática — pega de OTEL_RESOURCE_ATTRIBUTES, env vars, etc.
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		// Atributos explícitos (sobrepõem os detectados automaticamente)
		resource.WithAttributes(
			semconv.ServiceName(env.GetString(env.OTEL_SERVICE_NAME, "unknown-service")),
			semconv.ServiceVersion(env.GetString(env.OTEL_SERVICE_VERSION, "0.0.0")),
			semconv.DeploymentEnvironment(env.GetString(env.ENV, "dev")),
		),
	)

	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(
			metricExporter,
			metric.WithInterval(15*time.Second), // boa prática: não use só o timeout
			metric.WithTimeout(5*time.Second),
		)),
	)

	return meterProvider
}
