package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func InitMetrics() {}

func RegisterRelayGauges(activeChannels, totalSubscribers func() int) {
	m := otel.Meter("do-worker-relay")

	_, _ = m.Int64ObservableGauge("agentsmesh.relay.channels.active",
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(activeChannels()))
			return nil
		}))

	_, _ = m.Int64ObservableGauge("agentsmesh.relay.subscribers.active",
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(totalSubscribers()))
			return nil
		}))
}

// RegisterTunnelGauges wires the HTTP data-plane's live tunnel/stream counts
// (Registry.Stats) into OTel observable gauges, mirroring RegisterRelayGauges
// for the terminal data plane.
func RegisterTunnelGauges(activeTunnels, activeStreams func() int) {
	m := otel.Meter("do-worker-relay")

	_, _ = m.Int64ObservableGauge("gateway.tunnels.active",
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(activeTunnels()))
			return nil
		}))

	_, _ = m.Int64ObservableGauge("gateway.streams.active",
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(activeStreams()))
			return nil
		}))
}

// RecordPreviewRequest counts a completed /preview/* request by its final
// HTTP status. Instruments are looked up by name on every call rather than
// cached in a package-level singleton: preview requests are not a hot-enough
// path (unlike per-chunk body frames) to justify the caching complexity, and
// this keeps behavior correct across MeterProvider swaps (e.g. in tests).
func RecordPreviewRequest(ctx context.Context, status string) {
	c, err := otel.Meter("do-worker-relay").Int64Counter("gateway.preview.requests")
	if err != nil {
		return
	}
	c.Add(ctx, 1, metric.WithAttributes(attribute.String("status", status)))
}

// RecordPreviewBytes counts bytes flushed through a preview proxy stream,
// tagged by direction ("in" = browser->runner, "out" = runner->browser).
func RecordPreviewBytes(ctx context.Context, dir string, n int64) {
	if n <= 0 {
		return
	}
	c, err := otel.Meter("do-worker-relay").Int64Counter("gateway.preview.bytes")
	if err != nil {
		return
	}
	c.Add(ctx, n, metric.WithAttributes(attribute.String("dir", dir)))
}
