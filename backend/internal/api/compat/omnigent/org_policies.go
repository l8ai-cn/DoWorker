package omnigent

import (
	"fmt"
	"net/http"
	"strings"

	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

const acpToolRuleHandler = "acp_tool_rule"
const sessionCostBudgetHandler = "session_cost_budget"

func (d *Deps) handleListPolicies(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Policies == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := d.Policies.ListOrg(c.Request.Context(), tenant.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	data := make([]gin.H, 0, len(rows))
	for i := range rows {
		data = append(data, policyRowToWire(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": data})
}

func (d *Deps) handleCreatePolicy(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Policies == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body struct {
		Name          string         `json:"name"`
		Type          string         `json:"type"`
		Handler       string         `json:"handler"`
		FactoryParams map[string]any `json:"factory_params"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	in, err := parsePolicyInput(body.Handler, body.FactoryParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := d.Policies.CreateOrg(c.Request.Context(), tenant.OrganizationID, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	d.pushPolicyToActivePods(c)
	c.JSON(http.StatusOK, policyRowToWire(row))
}

func (d *Deps) handlePatchPolicy(c *gin.Context) {
	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Enabled != nil && !*body.Enabled {
		d.handleDeletePolicy(c)
		return
	}
	c.JSON(http.StatusNotImplemented, gin.H{"error": "only disable (delete) is supported"})
}

func (d *Deps) handleDeletePolicy(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Policies == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := permissionpolicysvc.ParsePolicyID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}
	if err := d.Policies.DeleteOrg(c.Request.Context(), tenant.OrganizationID, id); err != nil {
		if err == permissionpolicysvc.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	d.pushPolicyToActivePods(c)
	c.Status(http.StatusNoContent)
}

func policyRowToWire(row *permissionpolicysvc.OrgRow) gin.H {
	handler := row.PolicyHandler
	if handler == "" {
		handler = acpToolRuleHandler
	}
	params := gin.H{"priority": row.Priority}
	name := row.ToolPattern
	switch handler {
	case sessionCostBudgetHandler:
		if row.MaxUSD != nil {
			params["max_usd"] = *row.MaxUSD
		}
		name = "session_cost_budget"
	default:
		params["tool_pattern"] = row.ToolPattern
		params["verdict"] = row.Verdict
		if row.PathPattern != nil {
			params["path_pattern"] = *row.PathPattern
		}
		if row.AgentSlug != nil {
			params["agent_slug"] = *row.AgentSlug
		}
	}
	return gin.H{
		"id": fmt.Sprintf("pol_%d", row.ID), "object": "default_policy",
		"name": name, "type": "python", "handler": handler,
		"factory_params": params, "enabled": true,
		"created_at": row.CreatedAt.Unix(), "updated_at": row.UpdatedAt.Unix(),
		"created_by": nil,
	}
}

func parsePolicyInput(handler string, params map[string]any) (permissionpolicysvc.CreateInput, error) {
	switch handler {
	case acpToolRuleHandler, "":
		return parseToolRuleInput(handler, params)
	case sessionCostBudgetHandler:
		return parseCostBudgetInput(params)
	default:
		return permissionpolicysvc.CreateInput{}, fmt.Errorf("unsupported handler %q", handler)
	}
}

func parseCostBudgetInput(params map[string]any) (permissionpolicysvc.CreateInput, error) {
	max, ok := params["max_usd"].(float64)
	if !ok {
		if v, ok := params["max_usd"].(int); ok {
			max = float64(v)
		}
	}
	if max <= 0 {
		return permissionpolicysvc.CreateInput{}, fmt.Errorf("max_usd must be positive")
	}
	priority := 0
	switch p := params["priority"].(type) {
	case float64:
		priority = int(p)
	case int:
		priority = p
	}
	return permissionpolicysvc.CreateInput{
		PolicyHandler: permissionpolicysvc.HandlerSessionCostBudget,
		ToolPattern:   "*",
		Verdict:       "deny",
		Priority:      priority,
		MaxUSD:        &max,
	}, nil
}

func parseToolRuleInput(handler string, params map[string]any) (permissionpolicysvc.CreateInput, error) {
	if handler != "" && handler != acpToolRuleHandler {
		return permissionpolicysvc.CreateInput{}, fmt.Errorf("unsupported handler %q", handler)
	}
	tool, _ := params["tool_pattern"].(string)
	tool = strings.TrimSpace(tool)
	if tool == "" {
		return permissionpolicysvc.CreateInput{}, fmt.Errorf("tool_pattern required")
	}
	verdict, _ := params["verdict"].(string)
	verdict = strings.ToLower(strings.TrimSpace(verdict))
	if verdict != "allow" && verdict != "deny" && verdict != "ask" {
		return permissionpolicysvc.CreateInput{}, fmt.Errorf("verdict must be allow, deny, or ask")
	}
	in := permissionpolicysvc.CreateInput{PolicyHandler: permissionpolicysvc.HandlerACPToolRule, ToolPattern: tool, Verdict: verdict}
	if v, ok := params["path_pattern"].(string); ok && strings.TrimSpace(v) != "" {
		trimmed := strings.TrimSpace(v)
		in.PathPattern = &trimmed
	}
	if v, ok := params["agent_slug"].(string); ok && strings.TrimSpace(v) != "" {
		trimmed := strings.TrimSpace(v)
		in.AgentSlug = &trimmed
	}
	switch p := params["priority"].(type) {
	case float64:
		in.Priority = int(p)
	case int:
		in.Priority = p
	}
	return in, nil
}
