package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStripRelayPathPrefix(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"direct runner keeps /relay prefix", "/relay/runner/relay", "/runner/relay"},
		{"direct browser keeps /relay prefix", "/relay/browser/relay", "/browser/relay"},
		{"bare /relay maps to root", "/relay", "/"},
		{"already stripped by proxy", "/runner/relay", "/runner/relay"},
		{"health untouched", "/health", "/health"},
		{"lookalike not stripped", "/relayx/foo", "/relayx/foo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			h := stripRelayPathPrefix(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got = r.URL.Path
			}))
			req := httptest.NewRequest(http.MethodGet, tc.in, nil)
			h.ServeHTTP(httptest.NewRecorder(), req)
			if got != tc.want {
				t.Fatalf("path %q -> %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
