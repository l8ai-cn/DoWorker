package otel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// collectMetric finds a recorded instrument by name in a collected
// ResourceMetrics snapshot, across all scopes.
func collectMetric(t *testing.T, rm *metricdata.ResourceMetrics, name string) metricdata.Metrics {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m
			}
		}
	}
	t.Fatalf("metric %q not found in collected snapshot", name)
	return metricdata.Metrics{}
}

func sumAttr(t *testing.T, m metricdata.Metrics, key, value string) int64 {
	t.Helper()
	sum, ok := m.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("metric %q is not an int64 sum: %T", m.Name, m.Data)
	}
	for _, dp := range sum.DataPoints {
		v, ok := dp.Attributes.Value(attribute.Key(key))
		if ok && v.AsString() == value {
			return dp.Value
		}
	}
	t.Fatalf("no data point for %s=%s on metric %q", key, value, m.Name)
	return 0
}

func TestRegisterTunnelGauges_ReportsLiveCounts(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	RegisterTunnelGauges(func() int { return 3 }, func() int { return 7 })

	ctx := context.Background()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}

	tunnels := collectMetric(t, &rm, "gateway.tunnels.active")
	gauge, ok := tunnels.Data.(metricdata.Gauge[int64])
	if !ok || len(gauge.DataPoints) != 1 || gauge.DataPoints[0].Value != 3 {
		t.Fatalf("expected gateway.tunnels.active=3, got %#v", tunnels.Data)
	}

	streams := collectMetric(t, &rm, "gateway.streams.active")
	gauge, ok = streams.Data.(metricdata.Gauge[int64])
	if !ok || len(gauge.DataPoints) != 1 || gauge.DataPoints[0].Value != 7 {
		t.Fatalf("expected gateway.streams.active=7, got %#v", streams.Data)
	}
}

func TestPreviewCounters_RecordStatusAndBytes(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	ctx := context.Background()
	RecordPreviewRequest(ctx, "200")
	RecordPreviewRequest(ctx, "200")
	RecordPreviewRequest(ctx, "502")
	RecordPreviewBytes(ctx, "in", 128)
	RecordPreviewBytes(ctx, "out", 256)
	RecordPreviewBytes(ctx, "out", 64)

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}

	requests := collectMetric(t, &rm, "gateway.preview.requests")
	if got := sumAttr(t, requests, "status", "200"); got != 2 {
		t.Fatalf("expected 2 requests with status=200, got %d", got)
	}
	if got := sumAttr(t, requests, "status", "502"); got != 1 {
		t.Fatalf("expected 1 request with status=502, got %d", got)
	}

	bytes := collectMetric(t, &rm, "gateway.preview.bytes")
	if got := sumAttr(t, bytes, "dir", "in"); got != 128 {
		t.Fatalf("expected 128 bytes in, got %d", got)
	}
	if got := sumAttr(t, bytes, "dir", "out"); got != 320 {
		t.Fatalf("expected 320 bytes out, got %d", got)
	}
}
