package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	relaysvc "github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
)

const previewBootstrapTTL = 5 * time.Minute

// previewRelaySelector selects the relay/gateway edge for a pod. *relay.Manager
// satisfies it; tests inject a fake.
type previewRelaySelector interface {
	SelectRelayForPodGeo(opts relaysvc.GeoSelectOptions) *relaysvc.RelayInfo
}

// previewTokenGenerator mints typed relay tokens. *relay.TokenGenerator
// satisfies it; tests inject a fake.
type previewTokenGenerator interface {
	GenerateTypedToken(podKey string, runnerID, userID, orgID int64, tokenType, previewTarget string, expiry time.Duration) (string, error)
	GeneratePreviewBootstrapToken(podKey string, runnerID, userID, orgID int64, previewTarget, previewPath, previewOrigin string, expiry time.Duration) (string, error)
}

// GetPodPreview issues a short-lived Gateway session URL only after the Runner
// confirms its outbound tunnel is registered and ready.
//
// GET /api/v1/orgs/:slug/pods/:key/preview
func (h *PodHandler) GetPodPreview(c *gin.Context) {
	ctx := c.Request.Context()
	podKey := c.Param("key")

	pod, err := h.podService.GetPod(ctx, podKey)
	if err != nil {
		apierr.ResourceNotFound(c, "Pod not found")
		return
	}

	tenant := middleware.GetTenant(c)
	sub := policy.NewSubject(tenant.OrganizationID, tenant.UserID, tenant.UserRole)
	if !policy.PodPolicy.AllowRead(sub, h.podResourceWithGrants(ctx, podKey, pod.OrganizationID, pod.CreatedByID)) {
		apierr.ForbiddenAccess(c)
		return
	}

	route, err := relaysvc.ResolvePreviewRoute(pod)
	if err != nil {
		switch {
		case errors.Is(err, relaysvc.ErrPreviewDisabled):
			apierr.ResourceNotFound(c, "Preview is not enabled for this pod")
		case errors.Is(err, relaysvc.ErrPodNotActive):
			apierr.Conflict(c, "pod_not_active", "Pod is not active")
		default:
			apierr.InternalError(c, "failed to resolve preview route")
		}
		return
	}

	if h.relaySelector == nil || h.relayTokens == nil || h.commandSender == nil || h.previewPublicOrigin == "" {
		apierr.ServiceUnavailable(c, "preview_unavailable", "Preview is not available")
		return
	}

	relayInfo := h.relaySelector.SelectRelayForPodGeo(relaysvc.GeoSelectOptions{OrgSlug: tenant.OrganizationSlug})
	if relayInfo == nil {
		apierr.ServiceUnavailable(c, "no_relay", "No relay available")
		return
	}

	tunnelToken, err := h.relayTokens.GenerateTypedToken("", pod.RunnerID, 0, tenant.OrganizationID, "tunnel", "", 24*time.Hour)
	if err != nil {
		apierr.InternalError(c, "failed to mint tunnel token")
		return
	}
	if err := h.commandSender.SendConnectTunnel(ctx, pod.RunnerID, tunnelURLFromRelay(relayInfo.URL), tunnelToken); err != nil {
		apierr.ServiceUnavailable(c, "preview_unavailable", "Preview is not available")
		return
	}

	previewOrigin, err := previewOriginForPod(h.previewPublicOrigin, podKey)
	if err != nil {
		apierr.ServiceUnavailable(c, "preview_unavailable", "Preview is not available")
		return
	}
	previewToken, err := h.relayTokens.GeneratePreviewBootstrapToken(
		podKey,
		pod.RunnerID,
		tenant.UserID,
		tenant.OrganizationID,
		route.Target,
		route.Path,
		previewOrigin,
		previewBootstrapTTL,
	)
	if err != nil {
		apierr.InternalError(c, "failed to mint preview token")
		return
	}

	base := previewBaseURL(previewOrigin, podKey)
	c.JSON(http.StatusOK, gin.H{
		"preview_base_url": base,
		"session_url":      base + "__session?token=" + url.QueryEscape(previewToken),
		"expires_at":       time.Now().Add(previewBootstrapTTL).UTC().Format(time.RFC3339),
	})
}

func previewBaseURL(publicOrigin, podKey string) string {
	return fmt.Sprintf("%s/preview/%s/", strings.TrimRight(publicOrigin, "/"), url.PathEscape(podKey))
}

func previewOriginForPod(baseOrigin, podKey string) (string, error) {
	u, err := url.Parse(baseOrigin)
	if err != nil || u.Scheme == "" || u.Hostname() == "" || podKey == "" {
		return "", fmt.Errorf("invalid preview origin")
	}
	host := podKey + "." + u.Hostname()
	if port := u.Port(); port != "" {
		host += ":" + port
	}
	return u.Scheme + "://" + host, nil
}

// tunnelURLFromRelay derives the runner tunnel WebSocket URL from a relay's
// public WS URL: wss://host/relay -> wss://host/runner/tunnel.
func tunnelURLFromRelay(relayURL string) string {
	u, err := url.Parse(relayURL)
	if err != nil || u.Host == "" {
		return relayURL
	}
	return fmt.Sprintf("%s://%s/runner/tunnel", u.Scheme, u.Host)
}
