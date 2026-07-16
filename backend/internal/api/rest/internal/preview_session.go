package internal

import (
	"context"
	"errors"
	"net/http"
	"time"

	previewsvc "github.com/anthropics/agentsmesh/backend/internal/service/preview"
	"github.com/gin-gonic/gin"
)

const previewBootstrapTTL = 5 * time.Minute

type previewSessionService interface {
	Redeem(context.Context, string, previewsvc.SessionRecord, time.Duration) error
	Authorize(context.Context, previewsvc.SessionIdentity) error
}

type previewBootstrapRedeemRequest struct {
	BootstrapID string    `json:"bootstrap_id" binding:"required"`
	SessionID   string    `json:"session_id" binding:"required"`
	PodKey      string    `json:"pod_key" binding:"required"`
	UserID      int64     `json:"user_id" binding:"required"`
	OrgID       int64     `json:"org_id" binding:"required"`
	ExpiresAt   time.Time `json:"expires_at" binding:"required"`
}

type previewSessionAuthorizeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	PodKey    string `json:"pod_key" binding:"required"`
	UserID    int64  `json:"user_id" binding:"required"`
	OrgID     int64  `json:"org_id" binding:"required"`
}

func RegisterPreviewSessionRoutes(router *gin.RouterGroup, service previewSessionService) {
	router.POST("/preview-bootstrap/redeem", func(c *gin.Context) {
		if service == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "preview_session_unavailable"})
			return
		}
		var request previewBootstrapRedeemRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
		err := service.Redeem(c.Request.Context(), request.BootstrapID, previewsvc.SessionRecord{
			ID: request.SessionID, PodKey: request.PodKey, UserID: request.UserID,
			OrgID: request.OrgID, ExpiresAt: request.ExpiresAt,
		}, previewBootstrapTTL)
		writePreviewSessionResult(c, err, true)
	})
	router.POST("/preview-sessions/authorize", func(c *gin.Context) {
		if service == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "preview_session_unavailable"})
			return
		}
		var request previewSessionAuthorizeRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
		err := service.Authorize(c.Request.Context(), previewsvc.SessionIdentity{
			ID: request.SessionID, PodKey: request.PodKey, UserID: request.UserID, OrgID: request.OrgID,
		})
		writePreviewSessionResult(c, err, false)
	})
}

func writePreviewSessionResult(c *gin.Context, err error, redemption bool) {
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case redemption && errors.Is(err, previewsvc.ErrBootstrapConsumed):
		c.JSON(http.StatusConflict, gin.H{"error": "preview_bootstrap_consumed"})
	case errors.Is(err, previewsvc.ErrSessionUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "preview_session_unauthorized"})
	case errors.Is(err, previewsvc.ErrStoreUnavailable),
		errors.Is(err, previewsvc.ErrAuthorizationUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "preview_session_unavailable"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
	}
}
