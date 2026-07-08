package proxy

import (
	"net/http"
	"testing"
)

func TestSanitizeRequestHeaders(t *testing.T) {
	in := http.Header{
		"Connection":       {"keep-alive"},
		"Upgrade":          {"websocket"},
		"Proxy-Connection": {"x"},
		"X-Forwarded-For":  {"1.2.3.4"},
		"Content-Type":     {"text/html"},
	}
	out := SanitizeRequestHeaders(in, "9.9.9.9", "https", "host")
	if _, ok := out["Connection"]; ok {
		t.Fatal("hop-by-hop must be stripped")
	}
	if out.Get("Content-Type") != "text/html" {
		t.Fatal("passthrough header lost")
	}
	if out.Get("X-Forwarded-For") != "9.9.9.9" {
		t.Fatal("XFF must be rewritten, not trusted from client")
	}
	if out.Get("X-Forwarded-Proto") != "https" {
		t.Fatal("X-Forwarded-Proto not set")
	}
	if out.Get("X-Forwarded-Host") != "host" {
		t.Fatal("X-Forwarded-Host not set")
	}
}

func TestSanitizeRequestHeaders_ConnectionListedHeaderStripped(t *testing.T) {
	in := http.Header{
		"Connection":     {"X-Custom-Hop"},
		"X-Custom-Hop":   {"secret"},
		"Content-Length": {"5"},
	}
	out := SanitizeRequestHeaders(in, "1.1.1.1", "http", "h")
	if _, ok := out["X-Custom-Hop"]; ok {
		t.Fatal("header listed in Connection must be stripped")
	}
}

func TestSanitizeResponseHeaders(t *testing.T) {
	in := http.Header{
		"Connection":     {"keep-alive"},
		"Transfer-Encoding": {"chunked"},
		"Content-Type":   {"image/png"},
		"Content-Range":  {"bytes 0-1/2"},
		"Accept-Ranges":  {"bytes"},
	}
	out := SanitizeResponseHeaders(in)
	if _, ok := out["Connection"]; ok {
		t.Fatal("hop-by-hop must be stripped from response")
	}
	if _, ok := out["Transfer-Encoding"]; ok {
		t.Fatal("Transfer-Encoding must be stripped")
	}
	if out.Get("Content-Type") != "image/png" || out.Get("Content-Range") != "bytes 0-1/2" || out.Get("Accept-Ranges") != "bytes" {
		t.Fatal("content headers must be preserved")
	}
}
