package server

import (
	"fmt"
	"net/http"

	"github.com/anthropics/agentsmesh/relay/internal/config"
)

func (h *PreviewHandler) previewSessionCookie(
	podKey string,
	token string,
) (*http.Cookie, error) {
	cookie := &http.Cookie{
		Name:     previewCookieName,
		Value:    token,
		Path:     "/preview/" + podKey,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		MaxAge:   int(previewSessionTTL.Seconds()),
	}
	switch h.cfg.CookieMode {
	case config.PreviewCookieSameSite:
		cookie.SameSite = http.SameSiteStrictMode
	case config.PreviewCookiePartitioned:
		if !cookie.Secure {
			return nil, fmt.Errorf("partitioned preview cookie requires HTTPS")
		}
		cookie.SameSite = http.SameSiteNoneMode
		cookie.Partitioned = true
	default:
		return nil, fmt.Errorf("invalid preview cookie mode %q", h.cfg.CookieMode)
	}
	return cookie, nil
}
