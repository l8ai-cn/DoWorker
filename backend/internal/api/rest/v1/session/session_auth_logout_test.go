package sessionapi

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authservice "github.com/l8ai-cn/agentcloud/backend/internal/service/auth"
	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

type previewSessionRevokerStub struct {
	userID int64
}

func (s *previewSessionRevokerStub) RevokeUser(_ context.Context, userID int64) error {
	s.userID = userID
	return nil
}

func TestAuthLogoutRevokesPreviewSessions(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	const audience = "agentcloud-api"
	tokenManager, err := authpkg.NewAccessTokenManager(authpkg.AccessTokenConfig{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      "logout-test-key",
		Issuer:     "logout-test",
		Audiences:  []string{audience},
		Duration:   time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	token, err := tokenManager.GenerateToken(42, "user@example.com", "user", 1, "user")
	if err != nil {
		t.Fatal(err)
	}
	revoker := &previewSessionRevokerStub{}
	deps := Deps{
		Auth: authservice.NewService(&authservice.Config{
			AccessTokens:        tokenManager,
			AccessTokenAudience: audience,
		}, nil),
		PreviewSessions: revoker,
	}
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	ctx.Request.Header.Set("Authorization", "Bearer "+token)

	deps.handleAuthLogout(ctx)

	if ctx.Writer.Status() != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", ctx.Writer.Status())
	}
	if revoker.userID != 42 {
		t.Fatalf("revoked user = %d, want 42", revoker.userID)
	}
}
