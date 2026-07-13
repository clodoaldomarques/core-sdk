package tracer

import (
	"context"
	"encoding/json"

	"github.com/clodoaldomarques/core-sdk/internal/request"
	"github.com/clodoaldomarques/core-sdk/pkg/env"
	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	Cid     = "Cid"
	OrgID   = "OrgID"
	TraceID = "TraceID"
	Caller  = "Caller"
)

var (
	OtlService string
)

func init() {
	OtlService = env.GetString(env.OTEL_SERVICE_NAME, "")
}

type TraceSpan struct {
	s      trace.Span
	ctx    context.Context
	cid    string
	orgId  string
	hasErr bool
}

type Attributes map[string]any

func (t *TraceSpan) AddAttributes(a Attributes) {
	t.s.SetAttributes(buildKeyValue(a)...)
}

func (t *TraceSpan) AddEvent(n string, a Attributes) {
	t.s.AddEvent(n, trace.WithAttributes(buildKeyValue(a)...))
}

func (t *TraceSpan) SetError(e error) {
	t.s.RecordError(e)
	t.s.SetStatus(codes.Error, e.Error())
	t.hasErr = true
}

func (t *TraceSpan) SpanID() string {
	return t.s.SpanContext().SpanID().String()
}

func (t *TraceSpan) TraceID() string {
	return t.s.SpanContext().TraceID().String()
}

func (t *TraceSpan) AddEventAndLog(m string, a Attributes) {
	if a == nil {
		a = make(Attributes, 4)
	}

	a["Cid"] = t.cid
	a["OrgID"] = t.orgId
	a["TraceID"] = t.TraceID()
	a["SpanID"] = t.SpanID()

	t.AddEvent(m, a)
	logger.Info(t.ctx, m, logger.Fields(a))
}

func (t *TraceSpan) End() {
	if !t.hasErr {
		t.s.SetStatus(codes.Ok, "span finished without errors")
	}
	t.s.End()
}

func buildKeyValue(a Attributes) []attribute.KeyValue {
	values := make([]attribute.KeyValue, 0, len(a))
	for k, v := range a {
		switch t := v.(type) {
		case string:
			values = append(values, attribute.String(k, t))
			continue
		case int:
			values = append(values, attribute.Int(k, t))
			continue
		case int64:
			values = append(values, attribute.Int64(k, t))
			continue
		case bool:
			values = append(values, attribute.Bool(k, t))
			continue
		default:
			val, err := json.Marshal(v)
			if err != nil {
				continue
			}
			values = append(values, attribute.String(k, string(val)))
		}
	}
	return values
}

func NewSpanFromContext(ctx context.Context, spanName string, attributes ...attribute.KeyValue) (*TraceSpan, context.Context) {
	if spanName == "" {
		panic("spanName is required")
	}

	rc := request.GetRequestContext(ctx)
	cid := rc.Cid
	orgID := rc.OrgID

	attrs := []attribute.KeyValue{
		attribute.String(Cid, cid),
		attribute.String(OrgID, orgID),
	}
	attributes = append(attributes, attrs...)
	ctx, span := otel.Tracer(OtlService).Start(
		ctx,
		spanName,
		trace.WithAttributes(attributes...),
	)
	span.SetStatus(codes.Ok, spanName)
	return &TraceSpan{
		s:     span,
		ctx:   ctx,
		cid:   cid,
		orgId: orgID,
	}, ctx
}
