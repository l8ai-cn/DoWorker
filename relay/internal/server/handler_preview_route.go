package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
)

func (h *PreviewHandler) route(w http.ResponseWriter, r *http.Request) {
	if !h.requirePublicHost(w, r) {
		return
	}
	_, rest, ok := parsePreviewPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if rest == "__session" {
		h.handlePreviewSession(w, r)
		return
	}
	h.handlePreview(w, r)
}

func (h *PreviewHandler) requirePublicHost(w http.ResponseWriter, r *http.Request) bool {
	podKey, _, ok := parsePreviewPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return false
	}
	expectedHost, err := previewHostForPod(h.cfg.PublicHost, podKey)
	if err != nil || !strings.EqualFold(r.Host, expectedHost) {
		writePreviewError(w, "misdirected_request", http.StatusMisdirectedRequest)
		return false
	}
	return true
}

func previewOriginForPod(baseOrigin, podKey string) (string, error) {
	u, err := url.Parse(baseOrigin)
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return "", auth.ErrInvalidToken
	}
	host, err := previewHostForPod(u.Host, podKey)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + host, nil
}

func previewHostForPod(baseHost, podKey string) (string, error) {
	if baseHost == "" || podKey == "" {
		return "", auth.ErrInvalidToken
	}
	return podKey + "." + baseHost, nil
}

func parsePreviewPath(p string) (podKey, rest string, ok bool) {
	trimmed := strings.TrimPrefix(p, "/preview/")
	if trimmed == "" || trimmed == p {
		return "", "", false
	}
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		return trimmed[:idx], trimmed[idx+1:], true
	}
	return trimmed, "", true
}
