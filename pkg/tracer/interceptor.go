package tracer

import (
	"github.com/clodoaldomarques/core-sdk/internal/request"
	"github.com/labstack/echo/v4"
)

func Interceptor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := request.NewContext(c.Request().Context(), c.Request().Header, nil)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
