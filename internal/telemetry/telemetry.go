package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName    = "bgpin"
	serviceVersion = "0.1.0"
)

var (
	tracer trace.Tracer
)

// Config holds telemetry configuration
type Config struct {
	Enabled    bool
	ExportType string // stdout, otlp, jaeger
	Endpoint   string
}

// Initialize sets up OpenTelemetry
func Initialize(cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter based on config
	var exporter sdktrace.SpanExporter
	switch cfg.ExportType {
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	default:
		exporter, err = stdouttrace.New()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(serviceName)

	// Return shutdown function
	return tp.Shutdown, nil
}

// GetTracer returns the global tracer
func GetTracer() trace.Tracer {
	if tracer == nil {
		tracer = otel.Tracer(serviceName)
	}
	return tracer
}

// StartSpan starts a new span with common attributes
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name, trace.WithAttributes(attrs...))
}

// RecordLatency records latency metric on span
func RecordLatency(span trace.Span, start time.Time) {
	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("duration_ms", duration.Milliseconds()),
		attribute.String("duration", duration.String()),
	)
}

// RecordError records an error on span
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
	}
}

// RecordSuccess records success on span
func RecordSuccess(span trace.Span) {
	span.SetAttributes(attribute.Bool("success", true))
}

// Common attribute keys
var (
	AttrProvider   = attribute.Key("bgp.provider")
	AttrPrefix     = attribute.Key("bgp.prefix")
	AttrASN        = attribute.Key("bgp.asn")
	AttrASPath     = attribute.Key("bgp.as_path")
	AttrNextHop    = attribute.Key("bgp.next_hop")
	AttrCommand    = attribute.Key("cli.command")
	AttrOutputFmt  = attribute.Key("cli.output_format")
	AttrResultCnt  = attribute.Key("result.count")
)
