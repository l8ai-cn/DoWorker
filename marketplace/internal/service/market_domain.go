package service

import (
	"errors"
	"regexp"
	"strings"
)

var domainLabel = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func normalizePlatformBaseDomain(raw string) (string, error) {
	host := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(raw)), ".")
	labels := strings.Split(host, ".")
	if len(host) == 0 || len(host) > 253 || len(labels) < 2 {
		return "", errors.New("invalid marketplace public base domain")
	}
	for _, label := range labels {
		if !domainLabel.MatchString(label) {
			return "", errors.New("invalid marketplace public base domain")
		}
	}
	return host, nil
}
