// Package proxy implements the Gateway HTTP data-plane reverse proxy that
// forwards browser preview requests over a runner tunnel stream.
package proxy

import (
	"net/http"
	"strings"
)

// hopByHopHeaders are connection-specific headers that must never be forwarded
// end-to-end (RFC 7230 §6.1).
var hopByHopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Transfer-Encoding",
	"TE",
	"Trailer",
	"Upgrade",
}

// SanitizeRequestHeaders returns a copy of in with hop-by-hop headers and any
// client-supplied forwarding headers removed, then injects controlled
// X-Forwarded-* values derived from the trusted edge. The client can never
// spoof X-Forwarded-For/Proto/Host through this proxy.
func SanitizeRequestHeaders(in http.Header, clientIP, proto, host, hiddenCookieName string) http.Header {
	out := stripHopByHop(in)
	hideCookie(out, hiddenCookieName)
	out.Del("Referer")

	// Drop any inbound forwarding headers; we set our own authoritative values.
	for k := range out {
		ck := http.CanonicalHeaderKey(k)
		if ck == "Forwarded" || strings.HasPrefix(ck, "X-Forwarded-") {
			delete(out, k)
		}
	}

	if clientIP != "" {
		out.Set("X-Forwarded-For", clientIP)
	}
	if proto != "" {
		out.Set("X-Forwarded-Proto", proto)
	}
	if host != "" {
		out.Set("X-Forwarded-Host", host)
	}
	return out
}

// SanitizeResponseHeaders returns a copy of in with hop-by-hop headers removed.
// Content headers (Content-Type/Length/Range, Accept-Ranges, Content-Encoding,
// Cache-Control, etc.) are preserved so media/range/caching semantics survive.
func SanitizeResponseHeaders(in http.Header, hiddenCookieName string) http.Header {
	out := stripHopByHop(in)
	hideSetCookie(out, hiddenCookieName)
	return out
}

func hideCookie(header http.Header, hiddenName string) {
	if hiddenName == "" {
		return
	}
	var kept []string
	for _, line := range header.Values("Cookie") {
		for _, pair := range strings.Split(line, ";") {
			pair = strings.TrimSpace(pair)
			name, _, ok := strings.Cut(pair, "=")
			if ok && name != hiddenName {
				kept = append(kept, pair)
			}
		}
	}
	header.Del("Cookie")
	if len(kept) > 0 {
		header.Set("Cookie", strings.Join(kept, "; "))
	}
}

func hideSetCookie(header http.Header, hiddenName string) {
	if hiddenName == "" {
		return
	}
	values := header.Values("Set-Cookie")
	header.Del("Set-Cookie")
	for _, value := range values {
		first, _, _ := strings.Cut(value, ";")
		name, _, ok := strings.Cut(strings.TrimSpace(first), "=")
		if ok && name != hiddenName {
			header.Add("Set-Cookie", value)
		}
	}
}

// stripHopByHop clones in and removes hop-by-hop headers, including any header
// names explicitly listed in the Connection header.
func stripHopByHop(in http.Header) http.Header {
	out := in.Clone()
	if out == nil {
		out = http.Header{}
	}

	// Headers named by the Connection header are also hop-by-hop.
	for _, v := range in.Values("Connection") {
		for _, name := range strings.Split(v, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				out.Del(name)
			}
		}
	}

	for _, h := range hopByHopHeaders {
		out.Del(h)
	}
	return out
}
