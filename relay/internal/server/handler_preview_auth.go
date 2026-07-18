package server

import (
	"errors"
	"net/http"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	relaybackend "github.com/anthropics/agentsmesh/relay/internal/backend"
)

func (h *PreviewHandler) authenticate(
	w http.ResponseWriter,
	r *http.Request,
	podKey string,
) (*auth.RelayClaims, bool) {
	token := h.extractToken(r)
	if token == "" {
		writePreviewError(w, "token_required", http.StatusUnauthorized)
		return nil, false
	}
	expectedOrigin, err := previewOriginForPod(h.cfg.PublicOrigin, podKey)
	if err != nil {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return nil, false
	}
	claims, err := h.validator.ValidatePreviewToken(
		token,
		auth.TokenTypePreviewSession,
		expectedOrigin,
	)
	if err != nil || claims.PodKey != podKey {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return nil, false
	}
	if h.sessionBackend == nil {
		writePreviewError(w, "preview_session_unavailable", http.StatusServiceUnavailable)
		return nil, false
	}
	err = h.sessionBackend.AuthorizePreviewSession(
		r.Context(),
		previewSessionIdentity(claims),
	)
	if err == nil {
		return claims, true
	}
	if errors.Is(err, relaybackend.ErrPreviewSessionUnauthorized) {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
	} else {
		writePreviewError(w, "preview_session_unavailable", http.StatusServiceUnavailable)
	}
	return nil, false
}

func previewSessionIdentity(
	claims *auth.RelayClaims,
) relaybackend.PreviewSessionIdentity {
	return relaybackend.PreviewSessionIdentity{
		ID:     claims.ID,
		PodKey: claims.PodKey,
		UserID: claims.UserID,
		OrgID:  claims.OrgID,
	}
}
