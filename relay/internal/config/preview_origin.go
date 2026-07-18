package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func normalizePreviewPublicOrigin(raw, primaryDomain string, useHTTPS bool) (string, error) {
	origin, err := canonicalHTTPOrigin(raw)
	if err != nil {
		return "", fmt.Errorf("PREVIEW_PUBLIC_ORIGIN: %w", err)
	}
	if primaryDomain == "" {
		return origin, nil
	}
	scheme := "http"
	if useHTTPS {
		scheme = "https"
	}
	appOrigin, err := canonicalHTTPOrigin(scheme + "://" + primaryDomain)
	if err == nil && appOrigin == origin {
		return "", fmt.Errorf("PREVIEW_PUBLIC_ORIGIN must use a dedicated origin")
	}
	return origin, nil
}

func canonicalHTTPOrigin(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("must be an absolute http(s) origin: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if (scheme != "http" && scheme != "https") || u.Hostname() == "" {
		return "", fmt.Errorf("must be an absolute http(s) origin")
	}
	if u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", fmt.Errorf("must not contain credentials, path, query, or fragment")
	}
	hostname := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if hostname == "" || !isASCII(hostname) {
		return "", fmt.Errorf("host must use ASCII")
	}
	port := u.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		port = ""
	}
	host := hostname
	if strings.Contains(hostname, ":") {
		host = "[" + hostname + "]"
	}
	if port != "" {
		host = net.JoinHostPort(hostname, port)
	}
	return scheme + "://" + host, nil
}

func isASCII(value string) bool {
	for _, r := range value {
		if r > 127 {
			return false
		}
	}
	return true
}

func (c *Config) PreviewPublicHost() string {
	u, _ := url.Parse(c.PreviewPublicOrigin)
	return u.Host
}

func (c *Config) PreviewUsesHTTPS() bool {
	u, _ := url.Parse(c.PreviewPublicOrigin)
	return u.Scheme == "https"
}
