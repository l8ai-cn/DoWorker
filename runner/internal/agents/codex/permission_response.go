package codex

import (
	"fmt"
	"strconv"
)

const (
	requestUserInputMethod = "item/tool/requestUserInput"
	mcpElicitationMethod   = "mcpServer/elicitation/request"
)

func (t *transport) RespondToPermission(
	requestID string,
	approved bool,
	updatedInput map[string]any,
) error {
	rpcID, err := strconv.ParseInt(requestID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse request ID %q: %w", requestID, err)
	}
	method := t.takePermissionMethod(requestID)
	result := permissionResponse(method, approved, updatedInput)
	return t.tracker.Writer.WriteResponse(rpcID, result, nil)
}

func permissionResponse(
	method string,
	approved bool,
	updatedInput map[string]any,
) map[string]any {
	switch method {
	case requestUserInputMethod:
		return map[string]any{"answers": codexUserInputAnswers(updatedInput)}
	case mcpElicitationMethod:
		action := "decline"
		if approved {
			action = "accept"
		}
		return map[string]any{
			"action":  action,
			"content": mcpElicitationContent(updatedInput),
			"_meta":   nil,
		}
	default:
		decision := "decline"
		if approved {
			decision = "accept"
		}
		return map[string]any{"decision": decision}
	}
}

func mcpElicitationContent(updatedInput map[string]any) map[string]any {
	rawAnswers, ok := updatedInput["answers"].(map[string]any)
	if !ok {
		return updatedInput
	}
	result := make(map[string]any, len(rawAnswers))
	for field, raw := range rawAnswers {
		answers := stringAnswers(raw)
		if len(answers) == 1 {
			result[field] = answers[0]
		} else {
			result[field] = answers
		}
	}
	return result
}

func codexUserInputAnswers(updatedInput map[string]any) map[string]any {
	result := map[string]any{}
	rawAnswers, _ := updatedInput["answers"].(map[string]any)
	for questionID, raw := range rawAnswers {
		result[questionID] = map[string]any{"answers": stringAnswers(raw)}
	}
	return result
}

func stringAnswers(raw any) []string {
	switch value := raw.(type) {
	case string:
		return []string{value}
	case []string:
		return value
	case []any:
		result := make([]string, 0, len(value))
		for _, entry := range value {
			if answer, ok := entry.(string); ok {
				result = append(result, answer)
			}
		}
		return result
	default:
		return nil
	}
}

func (t *transport) rememberPermissionMethod(requestID, method string) {
	t.permissionMu.Lock()
	t.permissionMethods[requestID] = method
	t.permissionMu.Unlock()
}

func (t *transport) takePermissionMethod(requestID string) string {
	t.permissionMu.Lock()
	defer t.permissionMu.Unlock()
	method := t.permissionMethods[requestID]
	delete(t.permissionMethods, requestID)
	return method
}

func (t *transport) clearPermissionMethods() {
	t.permissionMu.Lock()
	clear(t.permissionMethods)
	t.permissionMu.Unlock()
}
