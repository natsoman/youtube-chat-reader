package oteltest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	t        *testing.T
	exporter *tracetest.InMemoryExporter
}

func NewTracer(t *testing.T) *Tracer {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()

	t.Cleanup(func() {
		_ = exporter.Shutdown(context.Background())
	})

	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	return &Tracer{t: t, exporter: exporter}
}

func (trc *Tracer) AssertSpan(name string, kind oteltrace.SpanKind, status trace.Status, attr ...attribute.KeyValue) {
	require.Len(trc.t, trc.exporter.GetSpans(), 1)
	actualSpan := trc.exporter.GetSpans()[0]
	assert.Equal(trc.t, name, actualSpan.Name)
	assert.Equal(trc.t, kind, actualSpan.SpanKind)
	assert.Equal(trc.t, status, actualSpan.Status)
	assert.ElementsMatch(trc.t, attr, actualSpan.Attributes)
}
