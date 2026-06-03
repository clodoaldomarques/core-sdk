package request

import "context"

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
	TraceID       string
	CustomHeaders map[string]any
}

func GetRequestContext(ctx context.Context) Context {
	if requestContext, ok := ctx.Value(ContextName).(Context); ok {
		return requestContext
	}
	return Context{}
}
