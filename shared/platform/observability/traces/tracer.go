package traces

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TracerInterface interface {
	Start(ctx context.Context, name string, options ...trace.SpanStartOption) (context.Context, trace.Span)
	RecordError(ctx context.Context, err error)
	End(span trace.Span, err error)
}

type Tracer struct {
	tracer trace.Tracer
}

func NewTracer(tracer trace.Tracer) TracerInterface {
	return &Tracer{tracer: tracer}
}

var (
	NotegicTracer TracerInterface
)

func (t *Tracer) Start(
	ctx context.Context,
	name string,
	options ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	ctx, span := t.tracer.Start(ctx, name, options...)
	caller := GetTrace(1)
	span.SetAttributes(
		attribute.String("code.filepath", caller.File),
		attribute.Int("code.lineno", caller.Line),
		attribute.String("code.function", caller.Function),
	)

	return ctx, span
}

func (t *Tracer) End(span trace.Span, err error) {
	if err != nil {
		t.recordError(span, err, 2)
	}

	span.End()
}

func (t *Tracer) RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	t.recordError(trace.SpanFromContext(ctx), err, 2)
}

func (t *Tracer) recordError(span trace.Span, err error, skip int) {
	caller := GetTrace(skip)
	span.SetStatus(codes.Error, err.Error())
	span.SetAttributes(
		attribute.String("error.type", fmt.Sprintf("%T", err)),
		attribute.String("error.source.filepath", caller.File),
		attribute.Int("error.source.lineno", caller.Line),
		attribute.String("error.source.function", caller.Function),
	)
	span.RecordError(err, trace.WithStackTrace(true))
	span.AddEvent("exception.stack", trace.WithAttributes(
		attribute.String("exception.stacktrace", stackString(GetTraces(skip, 16))),
	))
}

func stackString(items []Trace) string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.String())
	}

	return strings.Join(result, "\n")
}
