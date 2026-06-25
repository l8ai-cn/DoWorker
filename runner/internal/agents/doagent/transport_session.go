package doagent

import (
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/google/uuid"
)

const rpcTimeout = 30 * time.Second

func (t *transport) NewSession(cwd string, mcpServers map[string]any) (string, error) {
	sessionID := uuid.NewString()
	servers := acpFormatMCPServers(mcpServers)
	params := map[string]any{
		"sessionId":  sessionID,
		"cwd":        cwd,
		"mcpServers": servers,
	}

	pr, err := t.tracker.SendRequest("session/new", params)
	if err != nil {
		return "", fmt.Errorf("write session/new: %w", err)
	}

	resp, err := t.tracker.WaitResponse(pr, rpcTimeout)
	if err != nil {
		return "", fmt.Errorf("wait session/new response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("session/new error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
	}
	return sessionID, nil
}

func (t *transport) SendPrompt(sessionID, prompt string) error {
	params := map[string]any{
		"sessionId": sessionID,
		"prompt": []map[string]any{
			{"type": "text", "text": prompt},
		},
	}

	pr, err := t.tracker.SendRequest("session/prompt", params)
	if err != nil {
		return fmt.Errorf("write prompt: %w", err)
	}

	go func() {
		resp, err := t.tracker.WaitResponse(pr, 5*time.Minute)
		if err != nil {
			t.logger.Error("prompt response error", "error", err)
		} else if resp.Error != nil {
			t.logger.Error("prompt error", "code", resp.Error.Code, "message", resp.Error.Message)
		}
		if t.callbacks.OnStateChange != nil {
			t.callbacks.OnStateChange(acp.StateIdle)
		}
	}()

	return nil
}

func (t *transport) CancelSession(sessionID string) error {
	params := map[string]any{"sessionId": sessionID}
	return t.tracker.Writer.WriteNotification("session/cancel", params)
}

func acpFormatMCPServers(mcpServers map[string]any) []map[string]any {
	var servers []map[string]any
	for name, cfg := range mcpServers {
		entry := map[string]any{"name": name}
		if m, ok := cfg.(map[string]any); ok {
			for k, v := range m {
				entry[k] = v
			}
		}
		servers = append(servers, entry)
	}
	if servers == nil {
		servers = []map[string]any{}
	}
	return servers
}
