package sessionapi

import (
	"net/http"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
	"github.com/gin-gonic/gin"
)

type sessionRelayConnection struct {
	RelayURL string `json:"relay_url"`
	Token    string `json:"token"`
	PodKey   string `json:"pod_key"`
}

func (d *Deps) handleGetSessionRelayConnection(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	connection, status, message := d.sessionRelayConnection(c, pod)
	if status != 0 {
		c.JSON(status, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusOK, connection)
}

func (d *Deps) sessionRelayConnection(
	c *gin.Context,
	pod *podDomain.Pod,
) (*sessionRelayConnection, int, string) {
	if pod == nil || !pod.IsActive() || pod.RunnerID <= 0 {
		return nil, http.StatusNotFound, "session not found"
	}
	if pod.InteractionMode != podDomain.InteractionModePTY {
		return nil, http.StatusBadRequest, "terminal is only available for pty sessions"
	}
	if d.RelayManager == nil || d.RelayTokens == nil || d.CommandSender == nil {
		return nil, http.StatusServiceUnavailable, "relay unavailable"
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		return nil, http.StatusUnauthorized, "unauthorized"
	}
	subject := policy.NewSubject(tenant.OrganizationID, tenant.UserID, tenant.UserRole)
	resource := policy.PodResource(pod.OrganizationID, pod.CreatedByID)
	if !policy.PodPolicy.AllowRead(subject, resource) {
		return nil, http.StatusForbidden, "forbidden"
	}
	relayInfo := d.RelayManager.SelectRelayForPodGeo(relay.GeoSelectOptions{
		OrgSlug: tenant.OrganizationSlug,
	})
	if relayInfo == nil {
		return nil, http.StatusServiceUnavailable, "no relay"
	}
	runnerToken, err := d.RelayTokens.GenerateToken(
		pod.PodKey, pod.RunnerID, 0, tenant.OrganizationID, time.Hour,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, "token failed"
	}
	if err := d.CommandSender.SendSubscribePod(
		c.Request.Context(), pod.RunnerID, pod.PodKey, relayInfo.URL, runnerToken, true, 1000,
	); err != nil {
		return nil, http.StatusServiceUnavailable, "relay subscription failed"
	}
	browserToken, err := d.RelayTokens.GenerateToken(
		pod.PodKey, pod.RunnerID, tenant.UserID, tenant.OrganizationID, time.Hour,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, "token failed"
	}
	if browserToken == "" {
		return nil, http.StatusInternalServerError, "empty browser token"
	}
	return &sessionRelayConnection{
		RelayURL: relayInfo.URL,
		Token:    browserToken,
		PodKey:   pod.PodKey,
	}, 0, ""
}
