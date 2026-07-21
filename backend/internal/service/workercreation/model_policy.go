package workercreation

import (
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

var workerModelProtocolAdapters = map[string][]string{
	"do-agent":         {"openai-compatible", "anthropic"},
	"codex-cli":        {"openai-compatible"},
	"opencode":         {"openai-compatible"},
	"pattern-designer": {"openai-compatible"},
	"claude-code":      {"anthropic"},
	"gemini-cli":       {"gemini"},
	"minimax-cli":      {"minimax"},
	"openclaw":         {"openai-compatible"},
	"hermes":           {"openai-compatible"},
	"seedance-expert":  {"openai-compatible", "anthropic"},
	"video-studio":     {"openai-compatible"},
}

func validateWorkerModelRequirement(
	workerType slugkit.Slug,
	requirement specdomain.ModelRequirement,
) error {
	if !requirement.Required {
		return nil
	}
	expected, exists := workerModelProtocolAdapters[workerType.String()]
	if !exists {
		return nil
	}
	actual := modelProtocolAdapters(requirement.ProtocolAdapters)
	if !sameStringSet(actual, expected) {
		return invalidWorkerType("model protocol policy does not match definition")
	}
	return nil
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]int, len(left))
	for _, value := range left {
		seen[value]++
	}
	for _, value := range right {
		seen[value]--
		if seen[value] < 0 {
			return false
		}
	}
	return true
}
