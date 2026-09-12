package request

import (
	"context"
	"net/http"
)

const (
	HeaderXCid    = "x-cid"
	HeaderXTenant = "x-tenant"
	HeaderXCaller = "x-caller"
	ContextName   = "requestContext"
)

var (
	CustomHeaders = []string{"x-version"}
)

type Context struct {
	OrgID         string
	Cid           string
	Caller        string
	CustomHeaders map[string]any
}

func NewContext(ctx context.Context, h http.Header, customHeaders map[string]any) context.Context {
	return context.WithValue(ctx, ContextName, Context{
		OrgID:         h.Get(HeaderXTenant),
		Cid:           h.Get(HeaderXCid),
		Caller:        h.Get(HeaderXCaller),
		CustomHeaders: customHeaders,
	})
}

func GetRequestContext(ctx context.Context) Context {
	if requestContext, ok := ctx.Value(ContextName).(Context); ok {
		return requestContext
	}
	return Context{}
}
