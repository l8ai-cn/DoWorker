package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	relaybackend "github.com/anthropics/agentsmesh/relay/internal/backend"
)

const previewSessionTTL = 15 * time.Minute

func (h *PreviewHandler) HandlePreviewSession(w http.ResponseWriter, r *http.Request) {
	if !h.requirePublicHost(w, r) {
		return
	}
	h.handlePreviewSession(w, r)
}

func (h *PreviewHandler) handlePreviewSession(w http.ResponseWriter, r *http.Request) {
	podKey, _, ok := parsePreviewPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		writePreviewError(w, "token_required", http.StatusUnauthorized)
		return
	}
	expectedOrigin, err := previewOriginForPod(h.cfg.PublicOrigin, podKey)
	if err != nil {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return
	}
	claims, err := h.validator.ValidatePreviewToken(
		tokenStr,
		auth.TokenTypePreviewBootstrap,
		expectedOrigin,
	)
	if err != nil {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return
	}
	if claims.PodKey != podKey {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return
	}
	if h.sessionIssuer == nil {
		writePreviewError(w, "preview_session_unavailable", http.StatusServiceUnavailable)
		return
	}
	session, err := h.sessionIssuer.Issue(claims, previewSessionTTL)
	if err != nil {
		writePreviewError(w, "preview_session_unavailable", http.StatusServiceUnavailable)
		return
	}
	if h.sessionBackend == nil {
		writePreviewError(w, "preview_bootstrap_unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := h.sessionBackend.RedeemPreviewBootstrap(r.Context(), claims.ID, relaybackend.PreviewSessionRegistration{
		ID: session.ID, PodKey: claims.PodKey, UserID: claims.UserID,
		OrgID: claims.OrgID, ExpiresAt: session.ExpiresAt,
	}); err != nil {
		if errors.Is(err, relaybackend.ErrPreviewBootstrapConsumed) {
			writePreviewError(w, "invalid_token", http.StatusUnauthorized)
			return
		}
		writePreviewError(w, "preview_bootstrap_unavailable", http.StatusServiceUnavailable)
		return
	}

	cookie, err := h.previewSessionCookie(podKey, session.Token)
	if err != nil {
		writePreviewError(w, "preview_session_unavailable", http.StatusServiceUnavailable)
		return
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(w, r, "/preview/"+podKey+"/", http.StatusFound)
}
