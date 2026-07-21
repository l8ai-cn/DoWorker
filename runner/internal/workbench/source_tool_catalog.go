package workbench

import (
	"strings"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

type toolDescriptor struct {
	semanticKey string
	category    string
}

var commonTools = map[string]toolDescriptor{
	"Bash":                       {"shell.execute", "shell"},
	"shell":                      {"shell.execute", "shell"},
	"Read":                       {"filesystem.read", "filesystem"},
	"Write":                      {"filesystem.write", "filesystem"},
	"Edit":                       {"filesystem.edit", "filesystem"},
	"fileChange":                 {"filesystem.change", "filesystem"},
	"Grep":                       {"filesystem.search", "filesystem"},
	"WebFetch":                   {"web.fetch", "web"},
	"AskUserQuestion":            {"interaction.question", "interaction"},
	"image_generation":           {"media.image.generate", "media"},
	"web_search_call":            {"web.search", "web"},
	"file_search_call":           {"filesystem.search", "filesystem"},
	"computer_call":              {"computer.use", "computer"},
	"mcp_call":                   {"mcp.call", "mcp"},
	"mcp_list_tools":             {"mcp.list-tools", "mcp"},
	"code_interpreter_call":      {"code.interpret", "code"},
	"image_generation_call":      {"media.image.generate", "media"},
	"workbench.publish_artifact": {"artifact.publish", "artifact"},
}

func resolveToolIdentity(
	sourceProtocol, toolName string,
) (*agentworkbenchv2.ToolIdentity, string, bool) {
	descriptor, ok := commonTools[toolName]
	if !ok {
		return nil, "", false
	}
	namespace := "agentcloud." + sourceProtocol
	if toolName == "workbench.publish_artifact" {
		namespace = "agentcloud.runner"
	}
	return &agentworkbenchv2.ToolIdentity{
		Namespace:      namespace,
		SemanticKey:    descriptor.semanticKey,
		SchemaVersion:  "1",
		SourceToolName: stringPointer(toolName),
	}, descriptor.category, true
}

func sourceProtocol(adapterID string) string {
	switch {
	case strings.HasPrefix(adapterID, "codex"):
		return "codex"
	case strings.HasPrefix(adapterID, "claude"):
		return "claude"
	default:
		return "acp"
	}
}
