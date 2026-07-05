package doagent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/google/uuid"
)

const rpcTimeout = 30 * time.Second

func (t *transport) NewSession(cwd string, mcpServers map[string]any) (string, error) {
	return t.startSession("session/new", cwd, "", mcpServers)
}

func (t *transport) ResumeSession(cwd string, mcpServers map[string]any, externalSessionID string) (string, error) {
	return t.startSession("session/resume", cwd, externalSessionID, mcpServers)
}

func (t *transport) startSession(method, cwd, sessionID string, mcpServers map[string]any) (string, error) {
	servers := acpFormatMCPServers(mcpServers)
	params := map[string]any{
		"cwd":        cwd,
		"mcpServers": servers,
	}
	if sessionID != "" {
		params["sessionId"] = sessionID
	} else {
		params["sessionId"] = uuid.NewString()
	}

	pr, err := t.tracker.SendRequest(method, params)
	if err != nil {
		return "", fmt.Errorf("write %s: %w", method, err)
	}

	resp, err := t.tracker.WaitResponse(pr, rpcTimeout)
	if err != nil {
		return "", fmt.Errorf("wait %s response: %w", method, err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("%s error: code=%d msg=%s", method, resp.Error.Code, resp.Error.Message)
	}
	var result struct {
		SessionID string `json:"sessionId"`
	}
	if len(resp.Result) > 0 {
		_ = json.Unmarshal(resp.Result, &result)
	}
	if result.SessionID == "" {
		if sid, ok := params["sessionId"].(string); ok {
			result.SessionID = sid
		}
	}
	return result.SessionID, nil
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
