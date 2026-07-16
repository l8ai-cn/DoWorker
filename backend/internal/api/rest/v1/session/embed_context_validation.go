package sessionapi

import "net/url"

func validateEmbedOrigins(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, embedContextError("parent_origins is required")
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil ||
			(parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Path != "" ||
			parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, embedContextError("parent_origins must contain exact http or https origins")
		}
		origin := parsed.Scheme + "://" + parsed.Host
		if _, exists := seen[origin]; exists {
			return nil, embedContextError("parent_origins contains duplicates")
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	return result, nil
}

func validateEmbedCapabilities(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, embedContextError("capabilities is required")
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !supportedEmbedCapability(value) {
			return nil, embedContextError("capabilities contains an unsupported value")
		}
		if _, exists := seen[value]; exists {
			return nil, embedContextError("capabilities contains duplicates")
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	if _, ok := seen["read"]; !ok {
		return nil, embedContextError("read capability is required")
	}
	return result, nil
}

func supportedEmbedCapability(value string) bool {
	switch value {
	case "read", "write", "approve", "terminal", "control":
		return true
	default:
		return false
	}
}

type embedContextError string

func (e embedContextError) Error() string {
	return string(e)
}
