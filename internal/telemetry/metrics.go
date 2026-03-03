package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds all metric instruments
type Metrics struct {
	QueryCounter    metric.Int64Counter
	QueryDuration   metric.Float64Histogram
	ErrorCounter    metric.Int64Counter
	PrefixCounter   metric.Int64Counter
	NeighborCounter metric.Int64Counter
}

var globalMetrics *Metrics

// InitializeMetrics sets up metric instruments
func InitializeMetrics() error {
	meter := otel.GetMeterProvider().Meter(serviceName)

	queryCounter, err := meter.Int64Counter(
		"bgpin.queries.total",
		metric.WithDescription("Total number of BGP queries"),
		metric.WithUnit("{query}"),
	)
	if err != nil {
		return err
	}

	queryDuration, err := meter.Float64Histogram(
		"bgpin.query.duration",
		metric.WithDescription("Duration of BGP queries"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	errorCounter, err := meter.Int64Counter(
		"bgpin.errors.total",
		metric.WithDescription("Total number of errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	prefixCounter, err := meter.Int64Counter(
		"bgpin.prefixes.total",
		metric.WithDescription("Total number of prefixes queried"),
		metric.WithUnit("{prefix}"),
	)
	if err != nil {
		return err
	}

	neighborCounter, err := meter.Int64Counter(
		"bgpin.neighbors.total",
		metric.WithDescription("Total number of neighbors queried"),
		metric.WithUnit("{neighbor}"),
	)
	if err != nil {
		return err
	}

	globalMetrics = &Metrics{
		QueryCounter:    queryCounter,
		QueryDuration:   queryDuration,
		ErrorCounter:    errorCounter,
		PrefixCounter:   prefixCounter,
		NeighborCounter: neighborCounter,
	}

	return nil
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	return globalMetrics
}

// RecordQuery records a query metric
func RecordQuery(ctx context.Context, command string, duration time.Duration, success bool) {
	if globalMetrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("command", command),
		attribute.Bool("success", success),
	}

	globalMetrics.QueryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	globalMetrics.QueryDuration.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))

	if !success {
		globalMetrics.ErrorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordPrefixQuery records a prefix query
func RecordPrefixQuery(ctx context.Context, count int) {
	if globalMetrics == nil {
		return
	}

	globalMetrics.PrefixCounter.Add(ctx, int64(count))
}

// RecordNeighborQuery records a neighbor query
func RecordNeighborQuery(ctx context.Context, count int) {
	if globalMetrics == nil {
		return
	}

	globalMetrics.NeighborCounter.Add(ctx, int64(count))
}
