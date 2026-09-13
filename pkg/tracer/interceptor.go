package tracer

import (
	"github.com/clodoaldomarques/core-sdk/internal/request"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func Interceptor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// 1. Extrai trace context (traceparent/tracestate) dos headers.
			carrier := propagation.HeaderCarrier(req.Header)
			ctx := otel.GetTextMapPropagator().Extract(req.Context(), carrier)

			// 2. Popula request.Context (cid/org_id).
			ctx = request.NewContext(ctx, req.Header, nil)

			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}
