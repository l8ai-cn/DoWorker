package server

import (
	"net/http"
	"time"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
)

// HandlePreviewSession exchanges a one-shot preview JWT (query param) for a
// short-lived HttpOnly cookie scoped to this pod's preview path, then
// redirects to the preview base URL. This keeps the raw token out of every
// subsequent request URL/referrer once the iframe session is established.
//
// GET /preview/{podKey}/__session?token=<preview-jwt>
func (h *PreviewHandler) HandlePreviewSession(w http.ResponseWriter, r *http.Request) {
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
	claims, err := h.validator.ValidateToken(tokenStr)
	if err != nil {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return
	}
	if claims.ResolvedType() != auth.TokenTypePreview || claims.PodKey != podKey {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	maxAge := int(time.Until(claims.ExpiresAt.Time).Seconds())
	if maxAge <= 0 {
		writePreviewError(w, "token_expired", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     previewCookieName,
		Value:    tokenStr,
		Path:     "/preview/" + podKey,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
	http.Redirect(w, r, "/preview/"+podKey+"/", http.StatusFound)
}
