package sessionapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (d *Deps) handlePolicyRegistry(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data": []gin.H{
			{
				"handler":     sessionCostBudgetHandler,
				"kind":        "factory",
				"name":        "Session Cost Budget",
				"description": "Reject new turns when session spend exceeds max USD (pod stays alive)",
				"params_schema": gin.H{
					"type": "object",
					"properties": gin.H{
						"max_usd":  gin.H{"type": "number", "description": "Maximum session cost in USD"},
						"priority": gin.H{"type": "integer", "default": 0},
					},
					"required": []string{"max_usd"},
				},
			},
			{
				"handler":     "acp_tool_rule",
				"kind":        "factory",
				"name":        "ACP Tool Rule",
				"description": "Org-wide tool permission rule (allow / deny / ask)",
				"params_schema": gin.H{
					"type": "object",
					"properties": gin.H{
						"tool_pattern": gin.H{"type": "string", "description": "Tool name glob"},
						"path_pattern": gin.H{"type": "string", "description": "Optional path glob"},
						"verdict":      gin.H{"type": "string", "enum": []string{"allow", "deny", "ask"}},
						"priority":     gin.H{"type": "integer", "default": 0},
						"agent_slug":   gin.H{"type": "string", "description": "Optional agent scope"},
					},
					"required": []string{"tool_pattern", "verdict"},
				},
			},
		},
	})
}
