package codex

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func (t *transport) handleConfigWarning(params json.RawMessage) {
	var w struct {
		Summary string `json:"summary"`
		Details string `json:"details"`
	}
	if err := json.Unmarshal(params, &w); err != nil {
		return
	}
	msg := strings.TrimSpace(w.Summary)
	if msg == "" {
		msg = strings.TrimSpace(w.Details)
	}
	if msg == "" {
		msg = "codex config warning"
	}
	if t.callbacks.OnLog != nil {
		t.callbacks.OnLog("info", msg)
	}
}

func (t *transport) handleRawResponseItemCompleted(sid string, params json.RawMessage) {
	var body struct {
		Item json.RawMessage `json:"item"`
	}
	if err := json.Unmarshal(params, &body); err != nil || len(body.Item) == 0 {
		return
	}
	var item struct {
		Type    string                    `json:"type"`
		Text    string                    `json:"text"`
		Content []agentMessageContentPart `json:"content"`
	}
	if err := json.Unmarshal(body.Item, &item); err != nil {
		return
	}
	switch item.Type {
	case "agentMessage", "agent_message", "":
	default:
		return
	}
	emitAssistantChunk(t.callbacks, sid, agentMessageText(item.Text, item.Content))
}

func (t *transport) handleApprovalRequest(method string, rpcID int64, params json.RawMessage) {
	var req approvalRequestParams
	if err := json.Unmarshal(params, &req); err != nil {
		t.logger.Warn("failed to parse approval request", "method", method, "error", err)
		return
	}

	if t.callbacks.OnStateChange != nil {
		t.callbacks.OnStateChange(acp.StateWaitingPermission)
	}

	description := approvalDescription(method, req)
	toolName := approvalToolName(method, req)
	argsJSON := string(params)
	t.rememberPermissionMethod(fmt.Sprintf("%d", rpcID), method)

	if t.callbacks.OnPermissionRequest != nil {
		t.callbacks.OnPermissionRequest(acp.PermissionRequest{
			SessionID:     t.getSessionID(),
			RequestID:     fmt.Sprintf("%d", rpcID),
			ToolName:      toolName,
			Description:   description,
			ArgumentsJSON: argsJSON,
		})
	}
}

func approvalToolName(method string, req approvalRequestParams) string {
	switch method {
	case "item/fileChange/requestApproval":
		return "fileChange"
	case "item/permissions/requestApproval":
		return "permissions"
	case "item/tool/requestUserInput", "tool/requestUserInput":
		return "requestUserInput"
	case "mcpServer/elicitation/request":
		return "mcpElicitation"
	default:
		if req.Path != "" {
			return "fileChange"
		}
		return "command"
	}
}

func approvalDescription(method string, req approvalRequestParams) string {
	if d := strings.TrimSpace(req.Description); d != "" {
		return d
	}
	if d := strings.TrimSpace(req.Reason); d != "" {
		return d
	}
	if d := strings.TrimSpace(req.Command); d != "" {
		return d
	}
	if d := strings.TrimSpace(req.Path); d != "" {
		return d
	}
	if d := strings.TrimSpace(req.Message); d != "" {
		return d
	}
	return method
}
