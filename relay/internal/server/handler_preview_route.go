package server

import (
	"net/http"
	"strings"
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
	if h.cfg.PublicHost == "" || !strings.EqualFold(r.Host, h.cfg.PublicHost) {
		writePreviewError(w, "misdirected_request", http.StatusMisdirectedRequest)
		return false
	}
	return true
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
