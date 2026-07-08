package auth

import "strings"

// OriginChecker validates the WebSocket Origin header against an allowlist.
// An empty allowlist means allow-all (backward compatible with existing
// deployments where ALLOWED_ORIGINS is unset).
type OriginChecker struct {
	allowed  map[string]struct{}
	allowAll bool
}

// NewOriginChecker builds a checker from a list of allowed origins.
func NewOriginChecker(origins []string) *OriginChecker {
	oc := &OriginChecker{allowed: make(map[string]struct{})}
	n := 0
	for _, o := range origins {
		o = strings.TrimSpace(strings.ToLower(o))
		if o == "" {
			continue
		}
		oc.allowed[o] = struct{}{}
		n++
	}
	if n == 0 {
		oc.allowAll = true
	}
	return oc
}

// Allowed reports whether the given Origin is permitted. An empty origin
// (non-browser clients such as runners have no Origin header) is allowed;
// otherwise it must match the allowlist.
func (oc *OriginChecker) Allowed(origin string) bool {
	if origin == "" || oc.allowAll {
		return true
	}
	_, ok := oc.allowed[strings.ToLower(strings.TrimSpace(origin))]
	return ok
}
