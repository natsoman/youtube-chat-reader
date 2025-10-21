package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

type Telemetry struct {
	serviceName       string
	serviceVersion    string
	collectorGRPCAddr string

	sampleRate float64
	logLevel   slog.Level

	log       *slog.Logger
	shutdowns []func(ctx context.Context) error
}

// Configure configures OpenTelemetry tracing, metrics, and logging globally.
//
// After configuring, you can use [go.opentelemetry.io/otel.Tracer] and [go.opentelemetry.io/otel.Meter].
// To get an instrumented logger, use [slog.Default].
func Configure(ctx context.Context, serviceName, collectorGRPCAddr string, opts ...Option) (*Telemetry, error) {
	if collectorGRPCAddr == "" {
		return nil, errors.New("collector GRPC address is empty")
	}

	if serviceName == "" {
		return nil, errors.New("service name is empty")
	}

	otel.SetTextMapPropagator(propagation.TraceContext{})

	t := &Telemetry{
		serviceName:       serviceName,
		collectorGRPCAddr: collectorGRPCAddr,
		sampleRate:        1,
	}

	for _, opt := range opts {
		opt(t)
	}

	if err := t.configureTrace(ctx); err != nil {
		return nil, err
	}

	if err := t.configureSlog(ctx); err != nil {
		return nil, err
	}

	if err := t.configureMetric(ctx); err != nil {
		return nil, err
	}

	return t, nil
}

// Shutdown must be called before application exit to ensure all telemetry data is
// flushed and resources are properly released.
func (t *Telemetry) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, shutdown := range t.shutdowns {
		if err := shutdown(ctx); err != nil {
			fmt.Printf("Failed to shutdown OTEL: %v\n", err)
		}
	}
}

func (t *Telemetry) configureMetric(ctx context.Context) error {
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(t.collectorGRPCAddr),
	)
	if err != nil {
		return err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(
			exporter,
			metric.WithInterval(time.Second*5)),
		),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(t.serviceName),
			semconv.ServiceVersion(t.serviceVersion),
		)),
	)

	otel.SetMeterProvider(provider)

	t.shutdowns = append(t.shutdowns, provider.Shutdown)
	t.shutdowns = append(t.shutdowns, exporter.Shutdown)

	// Enable Go runtime metrics collection
	if err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second * 5)); err != nil {
		return err
	}

	return nil
}

func (t *Telemetry) configureSlog(ctx context.Context) error {
	exporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpoint(t.collectorGRPCAddr),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("create GRPC exporter: %w", err)
	}

	provider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(t.serviceName),
			semconv.ServiceVersion(t.serviceVersion),
		)),
	)

	global.SetLoggerProvider(provider)

	t.log = slog.New(newMultiHandler(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: t.logLevel}),
		otelslog.NewHandler(
			t.serviceName,
			otelslog.WithSource(true),
			otelslog.WithLoggerProvider(provider),
		),
	))

	slog.SetLogLoggerLevel(t.logLevel)
	slog.SetDefault(t.log)

	t.shutdowns = append(t.shutdowns, provider.Shutdown)
	t.shutdowns = append(t.shutdowns, exporter.Shutdown)

	return nil
}

func (t *Telemetry) configureTrace(ctx context.Context) error {
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(t.collectorGRPCAddr),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return err
	}

	provider := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(t.sampleRate)),
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(t.serviceName),
			semconv.ServiceVersion(t.serviceVersion),
		)),
	)

	otel.SetTracerProvider(provider)

	t.shutdowns = append(t.shutdowns, provider.Shutdown)
	t.shutdowns = append(t.shutdowns, exporter.Shutdown)

	return nil
}
