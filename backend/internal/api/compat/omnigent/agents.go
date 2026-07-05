package omnigent

import (
	"net/http"
	"os"

	"github.com/anthropics/agentsmesh/agentfile/capability"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type agentWire struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  *string           `json:"description,omitempty"`
	Harness      *string           `json:"harness,omitempty"`
	Skills       []skillWire       `json:"skills,omitempty"`
	Builtin      bool              `json:"builtin"`
	CreatedAt    int64             `json:"created_at"`
	Capabilities map[string]string `json:"capabilities,omitempty"`
}

type skillWire struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (d *Deps) handleListAgents(c *gin.Context) {
	_ = middleware.GetTenant(c)
	builtin, err := d.Agent.ListBuiltinAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list agents"})
		return
	}
	includeInternal := os.Getenv("AGENTSMESH_INCLUDE_INTERNAL_AGENTS") == "true"
	rows := make([]agentWire, 0, len(builtin))
	for _, a := range builtin {
		if !a.IsActive || (a.IsInternal && !includeInternal) {
			continue
		}
		harness := a.Slug
		if a.Executable != "" {
			harness = a.Executable
		}
		row := agentWire{
			ID:        a.Slug,
			Name:      a.Slug,
			Builtin:   a.IsBuiltin,
			CreatedAt: a.CreatedAt.Unix(),
			Harness:   &harness,
		}
		if a.Description != nil {
			row.Description = a.Description
		}
		if a.AgentfileSource != nil {
			row.Capabilities = capability.ScanDeclarations(*a.AgentfileSource)
		}
		rows = append(rows, row)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":     rows,
		"has_more": false,
		"last_id":  lastAgentID(rows),
	})
}

func lastAgentID(rows []agentWire) *string {
	if len(rows) == 0 {
		return nil
	}
	id := rows[len(rows)-1].ID
	return &id
}
