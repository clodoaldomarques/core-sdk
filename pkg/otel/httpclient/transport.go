package httpclient

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Transport struct {
	Base http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base

	if base == nil {
		base = http.DefaultTransport
	}

	ctx := req.Context()

	otel.GetTextMapPropagator().Inject(
		ctx,
		propagation.HeaderCarrier(req.Header),
	)

	return base.RoundTrip(req)
}
