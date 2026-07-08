package sessionapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var terminalUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (d *Deps) handleTerminalAttach(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if c.Param("terminal_id") != terminalMainID {
		c.JSON(http.StatusNotFound, gin.H{"error": "terminal not found", "code": "not_found"})
		return
	}
	if pod == nil || !pod.IsActive() || pod.RunnerID == 0 {
		closeTerminalWS(c, 4404, "session not found")
		return
	}
	if d.RelayManager == nil || d.RelayTokens == nil || d.CommandSender == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "relay unavailable"})
		return
	}
	tenant := middleware.GetTenant(c)
	sub := policy.NewSubject(tenant.OrganizationID, tenant.UserID, tenant.UserRole)
	if !policy.PodPolicy.AllowRead(sub, policy.PodResource(pod.OrganizationID, pod.CreatedByID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	relayInfo := d.RelayManager.SelectRelayForPodGeo(relay.GeoSelectOptions{OrgSlug: tenant.OrganizationSlug})
	if relayInfo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no relay"})
		return
	}
	ctx := c.Request.Context()
	runnerToken, err := d.RelayTokens.GenerateToken(pod.PodKey, pod.RunnerID, 0, tenant.OrganizationID, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token failed"})
		return
	}
	_ = d.CommandSender.SendSubscribePod(ctx, pod.RunnerID, pod.PodKey, relayInfo.URL, runnerToken, true, 1000)
	browserToken, err := d.RelayTokens.GenerateToken(pod.PodKey, pod.RunnerID, tenant.UserID, tenant.OrganizationID, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token failed"})
		return
	}
	clientWS, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	upstream, _, err := dialRelayWS(relayInfo.URL, browserToken)
	if err != nil {
		_ = clientWS.Close()
		return
	}
	_ = upstream.WriteMessage(websocket.BinaryMessage, []byte{relayMsgResync})
	readOnly := strings.EqualFold(c.Query("read_only"), "true")
	go func() {
		defer clientWS.Close()
		defer upstream.Close()
		bridgeTerminalWS(clientWS, upstream, readOnly)
	}()
	_ = row
}

func closeTerminalWS(c *gin.Context, code int, reason string) {
	ws, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": reason})
		return
	}
	_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason))
	_ = ws.Close()
}
