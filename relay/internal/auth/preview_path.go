package auth

import (
	"net/url"
	"path"
	"strings"
)

func NormalizePreviewPath(raw string) (string, error) {
	if raw == "" || strings.ContainsAny(raw, "?#") {
		return "", ErrInvalidToken
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil || !strings.HasPrefix(decoded, "/") {
		return "", ErrInvalidToken
	}
	for _, segment := range strings.Split(decoded, "/") {
		if segment == ".." {
			return "", ErrInvalidToken
		}
	}
	cleaned := path.Clean(decoded)
	return (&url.URL{Path: cleaned}).EscapedPath(), nil
}
