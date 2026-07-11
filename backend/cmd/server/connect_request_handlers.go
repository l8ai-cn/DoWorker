package main

import (
	"crypto/rand"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const connectPathPrefix = "/proto."

// Streaming Connect responses need per-write deadlines; the writer detects
// stream content types so unary RPCs retain the server's regular timeout.
func routeConnectOrREST(connectHandler, restHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, connectPathPrefix) {
			connectHandler.ServeHTTP(newStreamingResponseWriter(w), r)
			return
		}
		restHandler.ServeHTTP(w, r)
	})
}

func withConnectTracing(handler http.Handler) http.Handler {
	correlated := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if trace.SpanContextFromContext(request.Context()).IsValid() {
			handler.ServeHTTP(w, request)
			return
		}
		span, err := newRequestCorrelationSpan()
		if err != nil {
			http.Error(w, "request correlation unavailable", http.StatusInternalServerError)
			return
		}
		ctx := trace.ContextWithSpanContext(request.Context(), span)
		handler.ServeHTTP(w, request.WithContext(ctx))
	})
	return otelhttp.NewHandler(correlated, "connect.rpc", otelhttp.WithPropagators(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	))
}

func newRequestCorrelationSpan() (trace.SpanContext, error) {
	var traceID trace.TraceID
	if _, err := rand.Read(traceID[:]); err != nil {
		return trace.SpanContext{}, err
	}
	var spanID trace.SpanID
	if _, err := rand.Read(spanID[:]); err != nil {
		return trace.SpanContext{}, err
	}
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID}), nil
}
