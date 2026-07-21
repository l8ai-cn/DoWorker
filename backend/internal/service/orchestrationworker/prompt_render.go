package orchestrationworker

import (
	"sort"
	"strings"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

func renderPrompt(
	prompt resource.PromptSpec,
	inputs map[string]string,
) (string, error) {
	if len(validatePromptInputs(prompt, inputs)) != 0 {
		return "", control.ErrCorrupt
	}
	keys := make([]string, 0, len(prompt.Variables))
	for key := range prompt.Variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rendered := prompt.Content
	for _, key := range keys {
		value, exists := inputs[key]
		if !exists {
			if fallback := prompt.Variables[key].Default; fallback != nil {
				value = *fallback
			}
		}
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}
	return rendered, nil
}
