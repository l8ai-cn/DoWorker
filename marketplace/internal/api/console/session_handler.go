package console

import (
	"net/http"
	"strconv"

	actorapi "github.com/l8ai-cn/agentcloud/marketplace/internal/api/actor"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct{}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{}
}

func (h *SessionHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/me", h.getSession)
}

func (h *SessionHandler) getSession(c *gin.Context) {
	current, ok := actorapi.FromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTH_REQUIRED",
				"message": "请先登录后继续操作",
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":                  strconv.FormatInt(current.UserID, 10),
		"email":                    current.Email,
		"username":                 current.Username,
		"platform_organization_id": strconv.FormatInt(current.OrganizationID, 10),
		"platform_role":            current.Role,
	})
}
