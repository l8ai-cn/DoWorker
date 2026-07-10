package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestWithConnectTracingExtractsTraceparent(t *testing.T) {
	var got trace.TraceID
	handler := withConnectTracing(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		got = trace.SpanContextFromContext(request.Context()).TraceID()
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/proto.ai_resource.v1.AIResourceService/GetCatalog", nil)
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	handler.ServeHTTP(httptest.NewRecorder(), request)

	if got.String() != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace ID = %s", got.String())
	}
}

func TestWithConnectTracingGeneratesCorrelationWithoutExporterOrTraceparent(t *testing.T) {
	var got trace.TraceID
	handler := withConnectTracing(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		got = trace.SpanContextFromContext(request.Context()).TraceID()
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/proto.ai_resource.v1.AIResourceService/GetCatalog", nil)
	handler.ServeHTTP(httptest.NewRecorder(), request)

	if !got.IsValid() {
		t.Fatal("Connect request must have a valid correlation trace ID")
	}
}
