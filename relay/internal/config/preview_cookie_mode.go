package config

import (
	"fmt"
	"strings"
)

type PreviewCookieMode string

const (
	PreviewCookieSameSite    PreviewCookieMode = "same-site"
	PreviewCookiePartitioned PreviewCookieMode = "partitioned"
)

func normalizePreviewCookieMode(
	raw string,
	previewUsesHTTPS bool,
) (PreviewCookieMode, error) {
	mode := PreviewCookieMode(strings.TrimSpace(strings.ToLower(raw)))
	switch mode {
	case PreviewCookieSameSite:
		return mode, nil
	case PreviewCookiePartitioned:
		if !previewUsesHTTPS {
			return "", fmt.Errorf(
				"PREVIEW_COOKIE_MODE=partitioned requires an HTTPS preview origin",
			)
		}
		return mode, nil
	default:
		return "", fmt.Errorf(
			"PREVIEW_COOKIE_MODE must be same-site or partitioned",
		)
	}
}
