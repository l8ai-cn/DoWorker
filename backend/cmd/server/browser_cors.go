package main

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
)

func withBrowserCORS(allowedOrigins []string, handler http.Handler) http.Handler {
	allowedSet, wildcardAll := browserAllowedOrigins(allowedOrigins)
	return cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"Connect-Protocol-Version",
			"Connect-Timeout-Ms",
			"X-Organization-Slug",
			"X-API-Key",
		},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           600,
		AllowOriginFunc: func(origin string) bool {
			if wildcardAll {
				return true
			}
			if _, ok := allowedSet[origin]; ok {
				return true
			}
			return origin == "null" || strings.HasPrefix(origin, "file://")
		},
	}).Handler(handler)
}

func browserAllowedOrigins(allowedOrigins []string) (map[string]struct{}, bool) {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	allowedSet := make(map[string]struct{}, len(allowedOrigins))
	wildcardAll := false
	for _, origin := range allowedOrigins {
		allowedSet[origin] = struct{}{}
		wildcardAll = wildcardAll || origin == "*"
	}
	return allowedSet, wildcardAll
}
