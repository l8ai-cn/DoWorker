package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	imbridgesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/imbridge"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
)

type IMBridgeHandler struct {
	bridge *imbridgesvc.Bridge
}

func NewIMBridgeHandler(bridge *imbridgesvc.Bridge) *IMBridgeHandler {
	return &IMBridgeHandler{bridge: bridge}
}

func (h *IMBridgeHandler) ListProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"providers": h.bridge.ListProviders()})
}

func (h *IMBridgeHandler) ListConnections(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	items, err := h.bridge.ListConnections(c.Request.Context(), tenant.OrganizationID)
	if err != nil {
		apierr.InternalError(c, "Failed to list IM connections")
		return
	}
	c.JSON(http.StatusOK, gin.H{"connections": items})
}

func (h *IMBridgeHandler) GetConnection(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id, err := strconv.ParseInt(c.Param("connectionId"), 10, 64)
	if err != nil {
		apierr.ValidationError(c, "invalid connection id")
		return
	}
	row, err := h.bridge.GetConnection(c.Request.Context(), tenant.OrganizationID, id)
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"connection": row})
}

type createIMConnectionRequest struct {
	Provider  string          `json:"provider" binding:"required"`
	Name      string          `json:"name" binding:"required"`
	ChannelID *int64          `json:"channel_id"`
	Config    json.RawMessage `json:"config" binding:"required"`
	Status    string          `json:"status"`
}

func (h *IMBridgeHandler) CreateConnection(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req createIMConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	row, err := h.bridge.CreateConnection(c.Request.Context(), &imbridgesvc.CreateConnectionRequest{
		OrganizationID:  tenant.OrganizationID,
		CreatedByUserID: tenant.UserID,
		Provider:        req.Provider,
		Name:            req.Name,
		ChannelID:       req.ChannelID,
		Config:          req.Config,
		Status:          req.Status,
	})
	if err != nil {
		if errors.Is(err, imbridgesvc.ErrInvalidProvider) || errors.Is(err, imbridgesvc.ErrInvalidConfig) {
			apierr.ValidationError(c, err.Error())
			return
		}
		apierr.InternalError(c, "Failed to create IM connection")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"connection": row})
}

type updateIMConnectionRequest struct {
	Name      *string         `json:"name"`
	ChannelID *int64          `json:"channel_id"`
	Config    json.RawMessage `json:"config"`
	Status    *string         `json:"status"`
}

func (h *IMBridgeHandler) UpdateConnection(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id, err := strconv.ParseInt(c.Param("connectionId"), 10, 64)
	if err != nil {
		apierr.ValidationError(c, "invalid connection id")
		return
	}
	var req updateIMConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	row, err := h.bridge.UpdateConnection(c.Request.Context(), tenant.OrganizationID, id, &imbridgesvc.UpdateConnectionRequest{
		Name:      req.Name,
		ChannelID: req.ChannelID,
		Config:    req.Config,
		Status:    req.Status,
	})
	if err != nil {
		if errors.Is(err, imbridgesvc.ErrInvalidConfig) {
			apierr.ValidationError(c, err.Error())
			return
		}
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"connection": row})
}

func (h *IMBridgeHandler) DeleteConnection(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id, err := strconv.ParseInt(c.Param("connectionId"), 10, 64)
	if err != nil {
		apierr.ValidationError(c, "invalid connection id")
		return
	}
	if err := h.bridge.DeleteConnection(c.Request.Context(), tenant.OrganizationID, id); err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *IMBridgeHandler) notFoundOrInternal(c *gin.Context, err error) {
	if errors.Is(err, imbridgesvc.ErrNotFound) {
		apierr.ResourceNotFound(c, "IM connection not found")
		return
	}
	apierr.InternalError(c, "IM bridge request failed")
}

type startWeixinQRRequest struct {
	ConnectionID int64 `json:"connection_id" binding:"required"`
}

func (h *IMBridgeHandler) StartWeixinQRLogin(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req startWeixinQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	resp, err := h.bridge.StartWeixinQRLogin(c.Request.Context(), tenant.OrganizationID, req.ConnectionID)
	if err != nil {
		if errors.Is(err, imbridgesvc.ErrNotFound) {
			apierr.ResourceNotFound(c, "IM connection not found")
			return
		}
		apierr.ValidationError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *IMBridgeHandler) GetWeixinQRLoginStatus(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	sessionID := c.Param("sessionId")
	resp, err := h.bridge.PollWeixinQRLogin(c.Request.Context(), tenant.OrganizationID, sessionID)
	if err != nil {
		if errors.Is(err, imbridgesvc.ErrNotFound) {
			apierr.ResourceNotFound(c, "QR session not found")
			return
		}
		apierr.InternalError(c, "Weixin QR login failed")
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *IMBridgeHandler) GetWeixinQRImage(c *gin.Context) {
	sessionID := c.Param("sessionId")
	mediaType, data, err := h.bridge.GetWeixinQRImage(sessionID)
	if err != nil {
		if errors.Is(err, imbridgesvc.ErrNotFound) {
			apierr.ResourceNotFound(c, "QR session not found")
			return
		}
		apierr.InternalError(c, "QR image unavailable")
		return
	}
	c.Data(http.StatusOK, mediaType, data)
}
