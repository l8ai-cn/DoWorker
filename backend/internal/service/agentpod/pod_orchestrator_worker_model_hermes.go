package agentpod

import "fmt"

func hermesModelProvider(protocolAdapter string) (string, error) {
	switch protocolAdapter {
	case "openai-compatible":
		return "openai", nil
	case "anthropic", "gemini":
		return protocolAdapter, nil
	default:
		return "", fmt.Errorf("hermes does not support model protocol %q", protocolAdapter)
	}
}
