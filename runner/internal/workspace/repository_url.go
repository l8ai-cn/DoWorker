package workspace

import (
	"fmt"
	"net/url"
	"strings"
)

func validateRepositoryURL(raw string) error {
	if raw == "" {
		return nil
	}
	lower := strings.ToLower(raw)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid repository HTTP URL")
	}
	if parsed.Host == "" {
		return fmt.Errorf("repository HTTP URL requires a host")
	}
	if parsed.User != nil || strings.Contains(parsed.Host, "@") {
		return fmt.Errorf("repository HTTP URL must not contain userinfo")
	}
	if parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return fmt.Errorf("repository HTTP URL must not contain query or fragment")
	}
	return nil
}

func ValidateRepositoryURL(raw string) error {
	return validateRepositoryURL(raw)
}

func validateTokenRepositoryURL(raw string) error {
	if err := validateRepositoryURL(raw); err != nil {
		return err
	}
	parsed, err := url.Parse(raw)
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" {
		return fmt.Errorf("token authentication requires an HTTPS repository URL")
	}
	return nil
}

func validateRepositoryAuthURL(repoURL string, opts *WorktreeOptions) error {
	if opts != nil && opts.GitToken != "" {
		return validateTokenRepositoryURL(repoURL)
	}
	return validateRepositoryURL(repoURL)
}

func RepositoryURLForDisplay(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "[invalid repository URL]"
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return raw
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}
