package omnigent

import (
	"net/http"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/capability"
	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleGetSessionAgent(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Agent == nil {
		return
	}
	agent, err := d.Agent.GetBySlug(c.Request.Context(), row.AgentSlug)
	if err != nil || agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	harness := agent.Executable
	if harness == "" {
		harness = agent.Slug
	}
	mcpServers := domain.ParseMcpServers(row.McpServers)
	mcpWire := make([]gin.H, 0, len(mcpServers))
	for _, s := range mcpServers {
		mcpWire = append(mcpWire, mcpServerWire(s))
	}
	editable := agentSupportsMcp(agent)
	wire := gin.H{
		"id": agent.Slug, "object": "agent", "name": agent.Name,
		"harness": harness, "mcp_servers": mcpWire,
		"mcp_servers_editable": editable, "policies": []any{},
		"terminals": agentTerminals(agent),
	}
	if agent.Description != nil {
		wire["description"] = *agent.Description
	}
	if agent.AgentfileSource != nil {
		wire["capabilities"] = capability.ScanDeclarations(*agent.AgentfileSource)
	}
	c.JSON(http.StatusOK, wire)
}

func agentTerminals(agent *agentdomain.Agent) []string {
	if agent != nil && agent.SupportsMode("pty") {
		return []string{"shell"}
	}
	return []string{}
}

func agentSupportsMcp(agent *agentdomain.Agent) bool {
	if agent == nil || agent.AgentfileSource == nil {
		return false
	}
	caps := capability.ScanDeclarations(*agent.AgentfileSource)
	for _, c := range caps {
		if c == "mcp" || c == "MCP" {
			return true
		}
	}
	return strings.Contains(strings.ToLower(*agent.AgentfileSource), "mcp on")
}
