package logger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"github.com/labstack/echo/v4"
)

// Configuração opcional
type InterceptorConfig struct {
	MaxBodySize     int64    // bytes máximos a serem lidos para log (0 = ilimitado, cuidado!)
	RedactFields    []string // campos a ocultar no body (ex.: "password", "token")
	LogRequestBody  bool
	LogResponseBody bool
	LogHeaders      bool
}

var defaultConfig = InterceptorConfig{
	MaxBodySize:     10 * 1024, // 10KB
	LogRequestBody:  true,
	LogResponseBody: false, // geralmente desligado por segurança/performance
	LogHeaders:      false,
}

func Interceptor(next echo.HandlerFunc) echo.HandlerFunc {
	return InterceptorWithConfig(defaultConfig)(next)
}

func InterceptorWithConfig(cfg InterceptorConfig) echo.MiddlewareFunc {
	if cfg.MaxBodySize == 0 {
		cfg.MaxBodySize = 10 * 1024
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			ctx := req.Context()

			// 1. Lê e restaura o corpo da requisição (se configurado)
			var reqBody string
			if cfg.LogRequestBody && (req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch) {
				bodyBytes, err := readBodyWithLimit(req.Body, cfg.MaxBodySize)
				if err != nil {
					logger.Error(ctx, "failed to read request body", logger.Fields{"error": err.Error()})
				} else {
					reqBody = string(bodyBytes)
					// Restaura o body para o próximo handler
					req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
			}

			// 2. Prepara campos comuns do log
			fields := logger.Fields{
				"method": req.Method,
				"path":   c.Path(),
				"remote": c.RealIP(),
			}
			if reqBody != "" {
				fields["request_body"] = redactSensitive(reqBody, cfg.RedactFields)
			}
			if cfg.LogHeaders {
				fields["headers"] = filterHeaders(req.Header)
			}

			// Parâmetros da query string (GET) e path params
			if req.URL.RawQuery != "" {
				fields["query"] = req.URL.Query()
			}
			if len(c.ParamValues()) > 0 {
				fields["path_params"] = c.ParamValues()
			}

			logger.Info(ctx, "incoming request", fields)

			// 3. Capturar resposta (se configurado)
			var resBody bytes.Buffer
			resWriter := &bodyCaptureResponseWriter{
				ResponseWriter: c.Response().Writer,
				buf:            &resBody,
				maxSize:        cfg.MaxBodySize,
			}
			if cfg.LogResponseBody {
				c.Response().Writer = resWriter
			}

			// 4. Executar o próximo handler
			err := next(c)

			// 5. Coletar dados da resposta
			duration := time.Since(start)
			status := c.Response().Status
			logFields := logger.Fields{
				"status":      status,
				"duration_ms": duration.Milliseconds(),
			}

			if cfg.LogResponseBody && resWriter.written {
				responseBody := redactSensitive(resBody.String(), cfg.RedactFields)
				logFields["response_body"] = responseBody
			}

			// 6. Log do resultado final
			if err != nil {
				logFields["error"] = err.Error()
				logger.Error(ctx, "request failed", logFields)
			} else {
				if status >= 400 {
					logger.Warn(ctx, "request completed with client error", logFields)
				} else {
					logger.Info(ctx, "request completed successfully", logFields)
				}
			}

			return err
		}
	}
}

// Lê até N bytes do corpo, respeitando limite
func readBodyWithLimit(body io.ReadCloser, limit int64) ([]byte, error) {
	if body == nil {
		return []byte{}, nil
	}
	if limit <= 0 {
		return io.ReadAll(body)
	}
	return io.ReadAll(io.LimitReader(body, limit))
}

// ResponseWriter customizado que captura o corpo escrito
type bodyCaptureResponseWriter struct {
	http.ResponseWriter
	buf     *bytes.Buffer
	written bool
	maxSize int64
}

func (w *bodyCaptureResponseWriter) Write(b []byte) (int, error) {
	// Captura apenas até o limite
	if w.buf.Len() < int(w.maxSize) {
		remaining := w.maxSize - int64(w.buf.Len())
		copyLen := int(remaining)
		if copyLen > len(b) {
			copyLen = len(b)
		}
		w.buf.Write(b[:copyLen])
	}
	w.written = true
	return w.ResponseWriter.Write(b)
}

func (w *bodyCaptureResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

// Redação simples (apenas ilustrativa, use uma lib como 'github.com/elastic/go-elasticsearch/v8/estransform' ou regex)
func redactSensitive(body string, fields []string) string {
	if len(fields) == 0 {
		return body
	}
	// Exemplo simplificado: troca "password":"123" por "password":"***"
	// Na prática, use um parser JSON e redação recursiva
	// Aqui só para demonstrar a ideia
	return body // Implemente conforme necessidade
}

func filterHeaders(h http.Header) map[string]string {
	filtered := make(map[string]string)
	for k, v := range h {
		if k == "Authorization" {
			filtered[k] = "REDACTED"
		} else {
			filtered[k] = v[0] // pega apenas o primeiro valor
		}
	}
	return filtered
}
