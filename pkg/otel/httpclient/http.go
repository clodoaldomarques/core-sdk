package httpclient

import (
	"net/http"
)

func Client() *http.Client {
	return &http.Client{
		Transport: &Transport{
			Base: http.DefaultTransport,
		},
	}
}
