package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/l8ai-cn/agentcloud/relay/internal/protocol/tunnelframe"
	"github.com/l8ai-cn/agentcloud/relay/internal/tunnel"
)

func collectMetric(t *testing.T, rm *metricdata.ResourceMetrics, name string) (metricdata.Metrics, bool) {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m, true
			}
		}
	}
	return metricdata.Metrics{}, false
}

func TestProxyHTTP_RecordsRequestStatusAndByteMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	ft := newFakeTunnel()
	ft.onReqStart = func(st *tunnel.Stream, p tunnelframe.ReqStartPayload) {
		ft.inject(st.ID, tunnelframe.TypeRespStart, mustJSON(tunnelframe.RespStartPayload{Status: 200}))
		ft.inject(st.ID, tunnelframe.TypeRespBody, []byte("hello"))
		ft.inject(st.ID, tunnelframe.TypeRespEnd, nil)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/preview/pod1/x", nil)
	if err := ProxyHTTP(context.Background(), ft, rec, req, ProxyParams{
		PodKey: "pod1", Target: "127.0.0.1:3000", Path: "/x", WindowBytes: 1 << 20,
	}); err != nil {
		t.Fatal(err)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}

	requests, ok := collectMetric(t, &rm, "gateway.preview.requests")
	if !ok {
		t.Fatal("gateway.preview.requests not recorded")
	}
	sum, ok := requests.Data.(metricdata.Sum[int64])
	if !ok || len(sum.DataPoints) != 1 || sum.DataPoints[0].Value != 1 {
		t.Fatalf("expected one request recorded, got %#v", requests.Data)
	}

	bytesMetric, ok := collectMetric(t, &rm, "gateway.preview.bytes")
	if !ok {
		t.Fatal("gateway.preview.bytes not recorded")
	}
	byteSum, ok := bytesMetric.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("gateway.preview.bytes is not an int64 sum: %T", bytesMetric.Data)
	}
	var outBytes int64
	for _, dp := range byteSum.DataPoints {
		if v, ok := dp.Attributes.Value(attribute.Key("dir")); ok && v.AsString() == "out" {
			outBytes = dp.Value
		}
	}
	if outBytes != int64(len("hello")) {
		t.Fatalf("expected 5 bytes out, got %d", outBytes)
	}
}
