package expert

import (
	"errors"
	"strings"
)

var (
	ErrExpertNameRequired              = errors.New("expert name is required")
	ErrExpertAgentRequired             = errors.New("agent_slug is required")
	ErrExpertSnapshotUpdateUnsupported = errors.New(
		"expert runtime fields must be republished from a workerspec-backed pod",
	)
	ErrExpertSnapshotUnavailable = errors.New(
		"expert workerspec snapshot service is unavailable",
	)
)

func validateExpertBasics(agentSlug, name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrExpertNameRequired
	}
	if strings.TrimSpace(agentSlug) == "" {
		return ErrExpertAgentRequired
	}
	return nil
}

func normalizeInteractionMode(mode string) string {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case expertdomInteractionACP:
		return expertdomInteractionACP
	default:
		return expertdomInteractionPTY
	}
}

const (
	expertdomInteractionPTY = "pty"
	expertdomInteractionACP = "acp"
)

func trimOptional(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	return out
}
