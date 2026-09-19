package httpclient

import (
	"context"
	"net/http"
)

func Client(ctx context.Context) *http.Client {
	return &http.Client{
		Transport: &Transport{
			Base: http.DefaultTransport,
		},
	}
}
