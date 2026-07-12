package actor

import (
	"context"
	"net/http"
	"strings"

	authpkg "github.com/anthropics/agentsmesh/backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

const contextKey = "marketplace_actor"

type TokenVerifier interface {
	Verify(context.Context, string) (*authpkg.Claims, error)
}

type Actor struct {
	UserID         int64
	Email          string
	Username       string
	OrganizationID int64
	Role           string
}

func Middleware(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := parseBearer(c.GetHeader("Authorization"))
		if !ok {
			abortUnauthorized(c, "AUTH_REQUIRED", "请先登录后继续操作")
			return
		}
		claims, err := verifier.Verify(c.Request.Context(), token)
		if err != nil {
			abortUnauthorized(c, "INVALID_TOKEN", "登录状态无效或已过期")
			return
		}
		c.Set(contextKey, Actor{
			UserID:         claims.UserID,
			Email:          claims.Email,
			Username:       claims.Username,
			OrganizationID: claims.OrganizationID,
			Role:           claims.Role,
		})
		c.Next()
	}
}

func FromContext(c *gin.Context) (Actor, bool) {
	value, exists := c.Get(contextKey)
	if !exists {
		return Actor{}, false
	}
	current, ok := value.(Actor)
	return current, ok
}

func parseBearer(header string) (string, bool) {
	scheme, token, found := strings.Cut(header, " ")
	if !found || scheme != "Bearer" || strings.TrimSpace(token) == "" {
		return "", false
	}
	return token, true
}

func abortUnauthorized(c *gin.Context, code, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
