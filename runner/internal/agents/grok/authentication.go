package grok

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

const apiKeyAuthMethod = "xai.api_key"

type initializeResult struct {
	AuthMethods []authMethod `json:"authMethods"`
}

type authMethod struct {
	ID string `json:"id"`
}

func authenticate(requester acp.HandshakeRequester, initResult json.RawMessage) error {
	var result initializeResult
	if err := json.Unmarshal(initResult, &result); err != nil {
		return fmt.Errorf("parse initialize result: %w", err)
	}
	if os.Getenv("XAI_API_KEY") == "" {
		return fmt.Errorf("grok-build requires XAI_API_KEY for headless runner authentication")
	}
	if !hasAuthMethod(result.AuthMethods, apiKeyAuthMethod) {
		return fmt.Errorf("grok-build initialize result does not offer %s authentication", apiKeyAuthMethod)
	}
	_, err := requester.Request("authenticate", map[string]any{
		"methodId": apiKeyAuthMethod,
		"_meta": map[string]any{
			"headless": true,
		},
	})
	if err != nil {
		return fmt.Errorf("authenticate grok-build: %w", err)
	}
	return nil
}

func hasAuthMethod(methods []authMethod, want string) bool {
	for _, method := range methods {
		if method.ID == want {
			return true
		}
	}
	return false
}
