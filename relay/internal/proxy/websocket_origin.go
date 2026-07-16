package proxy

import (
	"net/http"
	"strings"
)

func previewWebSocketOriginAllowed(r *http.Request, expectedOrigin string) bool {
	return expectedOrigin != "" && strings.EqualFold(r.Header.Get("Origin"), expectedOrigin)
}
