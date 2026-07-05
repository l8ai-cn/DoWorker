package omnigent

import (
	"fmt"
	"net/http"

	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListSessionPolicies(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Policies == nil {
		return
	}
	rows, err := d.Policies.ListSession(c.Request.Context(), row.OrganizationID, row.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	data := make([]gin.H, 0, len(rows))
	for i := range rows {
		data = append(data, sessionPolicyRowToWire(&rows[i]))
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": data})
}

func (d *Deps) handleCreateSessionPolicy(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Policies == nil {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	var body struct {
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
	if in.PolicyHandler != permissionpolicysvc.HandlerACPToolRule {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only acp_tool_rule supported for session policies"})
		return
	}
	created, err := d.Policies.CreateSession(c.Request.Context(), row.OrganizationID, row.ID, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	d.pushPolicyToSessionPod(c, row, pod)
	c.JSON(http.StatusOK, sessionPolicyRowToWire(created))
}

func (d *Deps) handleDeleteSessionPolicy(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Policies == nil {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	id, err := permissionpolicysvc.ParsePolicyID(c.Param("policy_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}
	if err := d.Policies.DeleteSession(c.Request.Context(), row.OrganizationID, row.ID, id); err != nil {
		if err == permissionpolicysvc.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	d.pushPolicyToSessionPod(c, row, pod)
	c.Status(http.StatusNoContent)
}

func sessionPolicyRowToWire(row *permissionpolicysvc.OrgRow) gin.H {
	handler := row.PolicyHandler
	if handler == "" {
		handler = acpToolRuleHandler
	}
	params := gin.H{"priority": row.Priority, "tool_pattern": row.ToolPattern, "verdict": row.Verdict}
	if row.PathPattern != nil {
		params["path_pattern"] = *row.PathPattern
	}
	return gin.H{
		"id": fmt.Sprintf("pol_%d", row.ID), "object": "session.policy",
		"name": row.ToolPattern, "type": "python", "handler": handler,
		"factory_params": params, "enabled": true, "source": "session",
		"created_at": row.CreatedAt.Unix(), "updated_at": row.UpdatedAt.Unix(),
	}
}
