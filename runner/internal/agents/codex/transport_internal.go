package codex

import (
	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func (t *transport) dispatchMessage(msg *acp.JSONRPCMessage) {
	switch {
	case msg.IsResponse():
		t.tracker.HandleResponse(msg)
	case msg.IsNotification():
		t.handleNotification(msg.Method, msg.Params)
	case msg.IsRequest():
		if isApprovalRequest(msg.Method) {
			id, _ := msg.GetID()
			t.handleApprovalRequest(msg.Method, id, msg.Params)
		} else {
			t.tracker.RejectRequest(msg)
		}
	}
}

func isApprovalRequest(method string) bool {
	switch method {
	case "item/commandExecution/requestApproval",
		"item/fileChange/requestApproval",
		"item/permissions/requestApproval",
		"item/tool/requestUserInput",
		"tool/requestUserInput",
		"mcpServer/elicitation/request":
		return true
	default:
		return false
	}
}
