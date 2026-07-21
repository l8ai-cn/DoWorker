package codex

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func headlessAutomationFields(cwd string) map[string]any {
	params := map[string]any{
		"approvalPolicy": "never",
		"sandbox":        "danger-full-access",
	}
	if cwd != "" {
		params["cwd"] = cwd
	}
	return params
}

func mergeHeadlessFields(params map[string]any, cwd string) map[string]any {
	if params == nil {
		params = map[string]any{}
	}
	for key, value := range headlessAutomationFields(cwd) {
		params[key] = value
	}
	return params
}

func (t *transport) NewSession(cwd string, mcpServers map[string]any) (string, error) {
	t.workDir = cwd
	params := mergeHeadlessFields(nil, cwd)
	if mcpServers != nil {
		params["mcpServers"] = mcpServers
	}

	pr, err := t.tracker.SendRequest("thread/start", params)
	if err != nil {
		return "", fmt.Errorf("write thread/start: %w", err)
	}

	resp, err := t.tracker.WaitResponse(pr, 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("wait thread/start response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("thread/start error: code=%d msg=%s",
			resp.Error.Code, resp.Error.Message)
	}

	var result threadStartResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", fmt.Errorf("parse thread/start result: %w", err)
	}
	if err := t.activateSession(result.Thread.ID, result.Model); err != nil {
		return "", fmt.Errorf("activate thread/start result: %w", err)
	}

	return result.Thread.ID, nil
}

func (t *transport) ResumeSession(cwd string, mcpServers map[string]any, externalSessionID string) (string, error) {
	if externalSessionID == "" {
		return "", fmt.Errorf("thread id required")
	}
	t.workDir = cwd
	params := mergeHeadlessFields(map[string]any{"threadId": externalSessionID}, cwd)
	if mcpServers != nil {
		params["mcpServers"] = mcpServers
	}
	pr, err := t.tracker.SendRequest("thread/resume", params)
	if err != nil {
		return "", fmt.Errorf("write thread/resume: %w", err)
	}
	resp, err := t.tracker.WaitResponse(pr, 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("wait thread/resume response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("thread/resume error: code=%d msg=%s",
			resp.Error.Code, resp.Error.Message)
	}
	var result threadStartResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", fmt.Errorf("parse thread/resume result: %w", err)
	}
	if result.Thread.ID == "" {
		result.Thread.ID = externalSessionID
	}
	if err := t.activateSession(result.Thread.ID, result.Model); err != nil {
		return "", fmt.Errorf("activate thread/resume result: %w", err)
	}
	return result.Thread.ID, nil
}

func (t *transport) activateSession(sessionID, model string) error {
	if sessionID == "" {
		return fmt.Errorf("thread response missing id")
	}
	if err := t.setCurrentModel(model); err != nil {
		return err
	}
	t.sessionMu.Lock()
	t.sessionID = sessionID
	t.sessionMu.Unlock()
	if t.callbacks.OnConfigChange != nil {
		t.callbacks.OnConfigChange(sessionID, acp.ConfigUpdate{Model: model})
	}
	return nil
}
