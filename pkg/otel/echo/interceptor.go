package echo

import (
	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type InterceptorConfig struct {
	CountRequests  bool
	CountResponses bool
	SkipPaths      []string
}

var defaultConfig = InterceptorConfig{
	CountRequests:  true,
	CountResponses: true,
	SkipPaths: []string{
		"/health/live",
		"/health/ready",
	},
}

var (
	meterName = env.GetString(env.OTEL_SERVICE_NAME, "unknown-service")
)

func Interceptor() echo.MiddlewareFunc {
	return InterceptorWithConfig(defaultConfig)
}

func InterceptorWithConfig(config InterceptorConfig) echo.MiddlewareFunc {
	meter := otel.GetMeterProvider().Meter(meterName)

	requestsTotal, _ := meter.Int64Counter(
		"http.server.requests.total",
		metric.WithDescription("Total de requisições HTTP recebidas."),
		metric.WithUnit("{request}"), // Unidade recomendada para contadores
	)

	responsesTotal, _ := meter.Int64Counter(
		"http.server.responses.total",
		metric.WithDescription("Total de respostas HTTP enviadas."),
		metric.WithUnit("{response}"),
	)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if shouldSkipPath(c.Path(), config.SkipPaths) {
				return next(c)
			}

			attrs := []attribute.KeyValue{
				attribute.String("http.request.method", c.Request().Method),
				attribute.String("url.path", c.Path()),
			}

			if config.CountRequests {
				requestsTotal.Add(c.Request().Context(), 1, metric.WithAttributes(attrs...))
			}

			if config.CountResponses {
				responsesTotal.Add(c.Request().Context(), 1, metric.WithAttributes(attrs...))
			}

			err := next(c)

			responseAttrs := append(attrs,
				attribute.Int("http.response.status_code", c.Response().Status),
			)
			responsesTotal.Add(c.Request().Context(), 1, metric.WithAttributes(responseAttrs...))

			return err
		}
	}
}

func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}
