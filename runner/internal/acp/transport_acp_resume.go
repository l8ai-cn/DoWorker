package acp

import (
	"encoding/json"
	"fmt"
	"time"
)

func (t *ACPTransport) ResumeSession(cwd string, mcpServers map[string]any, externalSessionID string) (string, error) {
	if externalSessionID == "" {
		return "", fmt.Errorf("resume session id required")
	}
	servers := formatMCPServersArray(mcpServers)
	params := map[string]any{
		"sessionId":  externalSessionID,
		"cwd":        cwd,
		"mcpServers": servers,
	}
	pr, err := t.tracker.SendRequest("session/resume", params)
	if err != nil {
		return "", fmt.Errorf("write session/resume: %w", err)
	}
	resp, err := t.tracker.WaitResponse(pr, 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("wait session/resume response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("session/resume error: code=%d msg=%s",
			resp.Error.Code, resp.Error.Message)
	}
	var result struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", fmt.Errorf("parse session/resume result: %w", err)
	}
	if result.SessionID == "" {
		result.SessionID = externalSessionID
	}
	return result.SessionID, nil
}

func formatMCPServersArray(mcpServers map[string]any) []map[string]any {
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
