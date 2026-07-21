package v1

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
)

// Tests AuthHandler.isAllowedRedirect across the four redirect classes:
//   - same-host https (web SPA callback)
//   - relative path (in-app navigation)
//   - agentcloud:// deep link (Electron desktop OAuth callback)
//   - everything else (must be rejected)
func TestIsAllowedRedirect(t *testing.T) {
	h := &AuthHandler{config: &config.Config{PrimaryDomain: "agentcloud.ai"}}

	cases := []struct {
		name string
		url  string
		want bool
	}{
		{"same-host https", "https://agentcloud.ai/auth/callback", true},
		{"same-host with port", "https://agentcloud.ai:443/auth/callback", true},
		{"relative path", "/auth/callback", true},
		{"protocol-relative", "//evil.com/auth/callback", false},
		{"different host", "https://evil.com/auth/callback", false},
		{"http on prod host", "http://agentcloud.ai/auth/callback", true},
		{"desktop deep link", "agentcloud://oauth/callback", true},
		{"desktop deep link with query", "agentcloud://oauth/callback?token=abc", true},
		{"deep link wrong host", "agentcloud://attacker/oauth/callback", false},
		{"deep link wrong path", "agentcloud://oauth/evil", false},
		{"unknown scheme", "javascript:alert(1)", false},
		{"empty", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := h.isAllowedRedirect(tc.url)
			if got != tc.want {
				t.Errorf("isAllowedRedirect(%q) = %v, want %v", tc.url, got, tc.want)
			}
		})
	}
}
