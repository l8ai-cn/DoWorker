package omnigent

import (
	"net/http"

	"github.com/anthropics/agentsmesh/agentfile/capability"
	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
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
	wire := gin.H{
		"id": agent.Slug, "object": "agent", "name": agent.Name,
		"harness": harness, "mcp_servers": []any{},
		"mcp_servers_editable": false, "policies": []any{},
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
