package echo

import (
	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meterName = env.GetString(env.OTEL_SERVICE_NAME, "unknown-service")
)

// Middleware retorna um middleware customizado do Echo que registra
// contadores de requisições e respostas usando OpenTelemetry.
func Interceptor() echo.MiddlewareFunc {
	// Obtém o Meter do provider global (que sua SDK já configurou).
	meter := otel.GetMeterProvider().Meter(meterName)

	// Cria os contadores.
	// O nome segue a convenção semântica do OTel (sufixo _total para contadores).
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
			// Atributos comuns a serem usados em ambas as métricas.
			// Extraia o path da rota para evitar alta cardinalidade.
			// Use c.Path() para obter o padrão da rota (ex: /users/:id).
			attrs := []attribute.KeyValue{
				attribute.String("http.request.method", c.Request().Method),
				attribute.String("url.path", c.Path()),
			}

			// 1. Registra a requisição (antes de chamar o handler).
			requestsTotal.Add(c.Request().Context(), 1, metric.WithAttributes(attrs...))

			// Chama o próximo handler na cadeia.
			err := next(c)

			// 2. Registra a resposta (após o handler, com o status code).
			// Adiciona o status code como um atributo.
			responseAttrs := append(attrs,
				attribute.Int("http.response.status_code", c.Response().Status),
			)
			responsesTotal.Add(c.Request().Context(), 1, metric.WithAttributes(responseAttrs...))

			return err
		}
	}
}
