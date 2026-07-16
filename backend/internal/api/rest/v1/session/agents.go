package sessionapi

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/capability"
	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentpodservice "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/gin-gonic/gin"
)

type agentWire struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Description           *string           `json:"description,omitempty"`
	Harness               *string           `json:"harness,omitempty"`
	Skills                []skillWire       `json:"skills,omitempty"`
	Builtin               bool              `json:"builtin"`
	CreatedAt             int64             `json:"created_at"`
	Capabilities          map[string]string `json:"capabilities,omitempty"`
	SupportedModes        []string          `json:"supported_modes"`
	RequiresModelResource bool              `json:"requires_model_resource"`
}

type skillWire struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (d *Deps) handleListAgents(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant required"})
		return
	}
	if d.Agent == nil || d.Runner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "worker availability unavailable"})
		return
	}
	available, err := d.Runner.ListAvailableAgentSlugs(
		c.Request.Context(),
		tenant.OrganizationID,
		tenant.UserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list available workers"})
		return
	}
	builtin, err := d.Agent.ListBuiltinAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list agents"})
		return
	}
	includeInternal := os.Getenv("AGENTSMESH_INCLUDE_INTERNAL_AGENTS") == "true"
	rows, err := availableAgentRows(builtin, available, includeInternal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid agent interaction modes"})
		return
	}
	page, hasMore := paginateAgents(rows, c.Query("after"), agentPageSize)
	c.JSON(http.StatusOK, gin.H{
		"data":     page,
		"has_more": hasMore,
		"last_id":  lastAgentID(page),
	})
}

func availableAgentRows(
	builtin []*agentDomain.Agent,
	available []string,
	includeInternal bool,
) ([]agentWire, error) {
	availableSet := make(map[string]struct{}, len(available))
	for _, slug := range available {
		availableSet[slug] = struct{}{}
	}
	rows := make([]agentWire, 0, len(builtin))
	for _, a := range builtin {
		if !a.IsActive || (a.IsInternal && !includeInternal) {
			continue
		}
		if _, ok := availableSet[a.Slug]; !ok {
			continue
		}
		harness := a.Slug
		if a.Executable != "" {
			harness = a.Executable
		}
		modes, err := agentInteractionModes(a.SupportedModes)
		if err != nil {
			return nil, err
		}
		row := agentWire{
			ID:                    a.Slug,
			Name:                  a.Name,
			Builtin:               a.IsBuiltin,
			CreatedAt:             a.CreatedAt.Unix(),
			Harness:               &harness,
			SupportedModes:        modes,
			RequiresModelResource: agentpodservice.AgentRequiresModelResource(a),
		}
		if a.Description != nil {
			row.Description = a.Description
		}
		if a.AgentfileSource != nil {
			row.Capabilities = capability.ScanDeclarations(*a.AgentfileSource)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows, nil
}

func agentInteractionModes(raw string) ([]string, error) {
	modes := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, mode := range strings.Split(raw, ",") {
		mode = strings.TrimSpace(mode)
		if mode != "acp" && mode != "pty" {
			return nil, fmt.Errorf("unsupported interaction mode %q", mode)
		}
		if _, ok := seen[mode]; ok {
			return nil, fmt.Errorf("duplicate interaction mode %q", mode)
		}
		seen[mode] = struct{}{}
		modes = append(modes, mode)
	}
	if len(modes) == 0 {
		return nil, fmt.Errorf("no interaction modes configured")
	}
	return modes, nil
}

const agentPageSize = 50

func paginateAgents(rows []agentWire, after string, pageSize int) ([]agentWire, bool) {
	start := 0
	if after != "" {
		for i, r := range rows {
			if r.ID > after {
				start = i
				break
			}
			if i == len(rows)-1 {
				return nil, false
			}
		}
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], end < len(rows)
}

func lastAgentID(rows []agentWire) *string {
	if len(rows) == 0 {
		return nil
	}
	id := rows[len(rows)-1].ID
	return &id
}
