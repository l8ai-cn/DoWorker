package sessionapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	authservice "github.com/anthropics/agentsmesh/backend/internal/service/auth"
	orgservice "github.com/anthropics/agentsmesh/backend/internal/service/organization"
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(r gin.IRouter, d Deps) {
	auth := r.Group("/auth")
	auth.POST("/login", d.handleAuthLogin)
	auth.POST("/logout", d.handleAuthLogout)
	auth.GET("/me", accessTokenMiddleware(d), d.handleAuthMe)
}

func (d *Deps) handleAuthLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Username == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	if d.Auth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth service unavailable"})
		return
	}
	result, err := d.Auth.Login(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password."})
			return
		}
		if errors.Is(err, authservice.ErrUserDisabled) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed."})
		return
	}
	userID := result.User.Email
	if userID == "" {
		userID = result.User.Username
	}
	resp := gin.H{
		"user":       gin.H{"id": userID, "is_admin": result.User.IsSystemAdmin},
		"token":      result.Token,
		"expires_in": result.ExpiresIn,
	}
	if d.Org != nil {
		if orgs, err := d.Org.ListByUser(c.Request.Context(), result.User.ID); err == nil && len(orgs) > 0 {
			resp["org_slugs"] = orgservice.CollectOrgSlugs(orgs)
			resp["org_slug"] = orgservice.PickDefaultOrgSlug(orgs)
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (d *Deps) handleAuthLogout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token == "" || token == c.GetHeader("Authorization") {
		c.Status(http.StatusNoContent)
		return
	}
	if d.Auth == nil || d.PreviewSessions == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "logout service unavailable"})
		return
	}
	claims, err := d.Auth.ValidateTokenWithContext(c.Request.Context(), token)
	if err != nil {
		c.Status(http.StatusNoContent)
		return
	}
	if err := d.PreviewSessions.RevokeUser(c.Request.Context(), claims.UserID); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "preview session revocation failed"})
		return
	}
	if err := d.Auth.RevokeToken(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "token revocation failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleAuthMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}
	email, _ := c.Get("email")
	emailStr, _ := email.(string)
	if emailStr == "" {
		emailStr = "user"
	}
	isAdmin := false
	var createdAt *time.Time
	if d.User != nil {
		if u, err := d.User.GetByID(c.Request.Context(), userID); err == nil && u != nil {
			isAdmin = u.IsSystemAdmin
			if u.Email != "" {
				emailStr = u.Email
			}
			createdAt = &u.CreatedAt
		}
	}
	var createdUnix any
	if createdAt != nil {
		createdUnix = createdAt.Unix()
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            emailStr,
		"is_admin":      isAdmin,
		"created_at":    createdUnix,
		"last_login_at": nil,
	})
}
