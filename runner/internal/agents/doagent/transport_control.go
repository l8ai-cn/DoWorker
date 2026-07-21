package doagent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func (t *transport) RespondToPermission(requestID string, approved bool, _ map[string]any) error {
	decision := "deny"
	if approved {
		decision = "allow"
	}
	_, err := t.callRPC("permission/reply", map[string]any{
		"permissionId": requestID,
		"decision":     decision,
	})
	if err != nil {
		return fmt.Errorf("permission/reply: %w", err)
	}
	if t.callbacks.OnStateChange != nil {
		t.callbacks.OnStateChange(acp.StateProcessing)
	}
	return nil
}

func (t *transport) handlePermissionUpdated(params json.RawMessage) {
	var raw struct {
		SessionID  string `json:"sessionId"`
		Permission struct {
			ID    string          `json:"id"`
			Tool  string          `json:"tool"`
			Title string          `json:"title"`
			Input json.RawMessage `json:"input"`
		} `json:"permission"`
	}
	if err := json.Unmarshal(params, &raw); err != nil {
		t.logger.Warn("failed to parse permission.updated", "error", err)
		return
	}
	if t.callbacks.OnStateChange != nil {
		t.callbacks.OnStateChange(acp.StateWaitingPermission)
	}
	if t.callbacks.OnPermissionRequest == nil {
		return
	}
	desc := raw.Permission.Title
	if desc == "" {
		desc = raw.Permission.Tool
	}
	t.callbacks.OnPermissionRequest(acp.PermissionRequest{
		SessionID:   raw.SessionID,
		RequestID:   raw.Permission.ID,
		ToolName:    raw.Permission.Tool,
		Description: desc,
	})
}

func (t *transport) SendControlRequest(sessionID, subtype string, payload map[string]any) (map[string]any, error) {
	method, params, ok := mapControlToRPC(sessionID, subtype, payload)
	if !ok {
		return nil, acp.ErrControlNotSupported
	}
	return t.callRPC(method, params)
}

func mapControlToRPC(sessionID, subtype string, payload map[string]any) (string, map[string]any, bool) {
	switch subtype {
	case "set_model":
		model, _ := payload["model"].(string)
		return "session/setModel", map[string]any{"sessionId": sessionID, "model": model}, true
	case "set_execution_mode":
		mode, _ := payload["mode"].(string)
		if mode == "" {
			mode, _ = payload["modeId"].(string)
		}
		return "session/setMode", map[string]any{"sessionId": sessionID, "modeId": mode}, true
	case "doagent.rpc":
		method, _ := payload["method"].(string)
		if method == "" {
			return "", nil, false
		}
		params, _ := payload["params"].(map[string]any)
		if params == nil {
			params = map[string]any{}
		}
		if _, has := params["sessionId"]; !has && sessionID != "" {
			params["sessionId"] = sessionID
		}
		return method, params, true
	}
	if strings.HasPrefix(subtype, "goal/") || strings.HasPrefix(subtype, "session/") {
		params := payload
		if params == nil {
			params = map[string]any{}
		}
		if _, has := params["sessionId"]; !has && sessionID != "" {
			params["sessionId"] = sessionID
		}
		return subtype, params, true
	}
	return "", nil, false
}
