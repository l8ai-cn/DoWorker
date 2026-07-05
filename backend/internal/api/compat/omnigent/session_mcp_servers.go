package omnigent

import (
	"net/http"
	"strings"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleCreateMcpServer(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	var body domain.McpServer
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	servers := domain.ParseMcpServers(row.McpServers)
	for _, s := range servers {
		if s.Name == body.Name {
			c.JSON(http.StatusConflict, gin.H{"error": "server already exists"})
			return
		}
	}
	servers = append(servers, body)
	if err := d.Sessions.SetMcpServers(c.Request.Context(), row.ID, servers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	row.McpServers, _ = marshalJSON(servers)
	if _, err := d.rebuildSessionPod(c, row, pod, row.AgentSlug); err != nil {
		writeSwitchAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, mcpServerWire(body))
}

func (d *Deps) handleUpdateMcpServer(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	name := c.Param("server_name")
	var body domain.McpServer
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	servers := domain.ParseMcpServers(row.McpServers)
	found := false
	for i := range servers {
		if servers[i].Name == name {
			body.Name = name
			servers[i] = body
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := d.Sessions.SetMcpServers(c.Request.Context(), row.ID, servers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	row.McpServers, _ = marshalJSON(servers)
	if _, err := d.rebuildSessionPod(c, row, pod, row.AgentSlug); err != nil {
		writeSwitchAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, mcpServerWire(body))
}

func (d *Deps) handleDeleteMcpServer(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	name := c.Param("server_name")
	servers := domain.ParseMcpServers(row.McpServers)
	next := make([]domain.McpServer, 0, len(servers))
	found := false
	for _, s := range servers {
		if s.Name == name {
			found = true
			continue
		}
		next = append(next, s)
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := d.Sessions.SetMcpServers(c.Request.Context(), row.ID, next); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	row.McpServers, _ = marshalJSON(next)
	if _, err := d.rebuildSessionPod(c, row, pod, row.AgentSlug); err != nil {
		writeSwitchAgentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func mcpServerWire(s domain.McpServer) gin.H {
	return gin.H{
		"name": s.Name, "transport": s.Transport,
		"description": s.Description, "url": s.URL,
		"command": s.Command, "args": s.Args,
	}
}
