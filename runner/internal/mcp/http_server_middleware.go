package mcp

import (
	"fmt"
)

// registerTools registers all collaboration tools.
func (s *HTTPServer) registerTools() {
	s.tools = []*MCPTool{
		// Pod interaction tools
		s.createGetPodSnapshotTool(),
		s.createSendPodInputTool(),
		s.createGetPodStatusTool(),
		s.createWorkbenchPublishArtifactTool(),

		// Discovery tools
		s.createListAvailablePodsTool(),
		s.createListRunnersTool(),
		s.createListRepositoriesTool(),

		// Binding tools
		s.createBindPodTool(),
		s.createAcceptBindingTool(),
		s.createRejectBindingTool(),
		s.createUnbindPodTool(),
		s.createGetBindingsTool(),
		s.createGetBoundPodsTool(),

		// Channel tools
		s.createSearchChannelsTool(),
		s.createCreateChannelTool(),
		s.createGetChannelTool(),
		s.createSendChannelMessageTool(),
		s.createGetChannelMessagesTool(),
		s.createGetChannelDocumentTool(),
		s.createUpdateChannelDocumentTool(),

		// Ticket tools
		s.createSearchTicketsTool(),
		s.createGetTicketTool(),
		s.createCreateTicketTool(),
		s.createUpdateTicketTool(),
		s.createDeleteTicketTool(),
		s.createPostCommentTool(),

		// Pod tools
		s.createCreatePodTool(),

		// Workflow tools
		s.createListWorkflowsTool(),
		s.createCreateWorkflowTool(),
		s.createTriggerWorkflowTool(),

		// Block Store tools — structured collaboration substrate (notes,
		// tasks, views, indicators, triggers). See http_tools_block.go.
		s.createBlockCreateTool(),
		s.createBlockUpdateTool(),
		s.createBlockDeleteTool(),
		s.createBlockAddRefTool(),
		s.createBlockRemoveRefTool(),
		s.createBlockUpdateRefTool(),
		s.createIndicatorDefineTool(),
		s.createTriggerDefineTool(),
		s.createMemoryRetrieveTool(),
		s.createBlockListTypesTool(),
		s.createBlockListWorkspacesTool(),
		s.createBlockGetDefaultWorkspaceTool(),

		// Knowledge Base tools — llm-wiki workflow. See http_tools_kb.go.
		s.createKbListTool(),
		s.createKbSearchTool(),
		s.createKbReadTool(),
		s.createKbWriteTool(),
	}
}

// GenerateMCPConfig generates the MCP configuration JSON for Claude Code.
func (s *HTTPServer) GenerateMCPConfig(podKey string) map[string]interface{} {
	return map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"agentcloud-collaboration": map[string]interface{}{
				"command": "curl",
				"args": []string{
					"-X", "POST",
					"-H", "Content-Type: application/json",
					"-H", fmt.Sprintf("X-Pod-Key: %s", podKey),
					fmt.Sprintf("http://localhost:%d/mcp", s.port),
					"-d", "@-",
				},
			},
		},
	}
}
