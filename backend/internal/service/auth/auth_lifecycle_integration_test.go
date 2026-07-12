package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	userService "github.com/anthropics/agentsmesh/backend/internal/service/user"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAuthService creates an auth service backed by real SQLite + miniredis.
func newTestAuthService(t *testing.T) (*Service, *userService.Service) {
	t.Helper()

	db := testkit.SetupTestDB(t)
	userRepo := infra.NewUserRepository(db)
	userSvc := userService.NewService(userRepo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { redisClient.Close() })

	cfg := &Config{
		JWTExpiration:     time.Hour,
		RefreshExpiration: 24 * time.Hour,
		Issuer:            "test-issuer",
	}
	configureTestAccessTokens(t, cfg)

	authSvc := NewServiceWithRedis(cfg, userSvc, redisClient)
	return authSvc, userSvc
}

// createTestUser registers a user through the user service for auth tests.
func createTestUser(t *testing.T, userSvc *userService.Service, email, password string) string {
	t.Helper()
	ctx := context.Background()
	local := strings.SplitN(email, "@", 2)[0]
	username := slugkit.Sanitize(local)
	if len(username) < slugkit.MinLen {
		username = "testuser-" + strings.ReplaceAll(slugkit.Sanitize(email), ".", "-")
	}
	_, err := userSvc.Create(ctx, &userService.CreateRequest{
		Email:    email,
		Username: username,
		Name:     "Test User",
		Password: password,
	})
	require.NoError(t, err)
	return username
}

func TestAuth_LoginSuccess(t *testing.T) {
	authSvc, userSvc := newTestAuthService(t)
	ctx := context.Background()

	username := createTestUser(t, userSvc, "alice@example.com", "correctpassword")

	result, err := authSvc.Login(ctx, username, "correctpassword")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.Token, "access token should be returned")
	assert.NotEmpty(t, result.RefreshToken, "refresh token should be returned")
	assert.Equal(t, "alice@example.com", result.User.Email)
	assert.Greater(t, result.ExpiresIn, int64(0))
}

func TestAuth_LoginInvalidPassword(t *testing.T) {
	authSvc, userSvc := newTestAuthService(t)
	ctx := context.Background()

	username := createTestUser(t, userSvc, "bob@example.com", "realpassword")

	result, err := authSvc.Login(ctx, username, "wrongpassword")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuth_LoginNonexistentUser(t *testing.T) {
	authSvc, _ := newTestAuthService(t)
	ctx := context.Background()

	result, err := authSvc.Login(ctx, "ghost", "whatever")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuth_TokenGeneration(t *testing.T) {
	authSvc, userSvc := newTestAuthService(t)
	ctx := context.Background()

	createTestUser(t, userSvc, "carol@example.com", "pass123")
	u, err := userSvc.GetByEmail(ctx, "carol@example.com")
	require.NoError(t, err)

	pair, err := authSvc.GenerateTokenPairWithContext(ctx, u, 42, "admin")
	require.NoError(t, err)
	require.NotNil(t, pair)

	// Parse the JWT and verify claims
	claims, err := authSvc.ValidateToken(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, u.ID, claims.UserID)
	assert.Equal(t, "carol@example.com", claims.Email)
	assert.Equal(t, int64(42), claims.OrganizationID)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "test-issuer", claims.Issuer)
}

func TestAuth_RegisterSuccess(t *testing.T) {
	authSvc, _ := newTestAuthService(t)
	ctx := context.Background()

	result, err := authSvc.Register(ctx, &RegisterRequest{
		Email:    "newuser@example.com",
		Username: "newuser",
		Password: "securePass1",
		Name:     "New User",
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "newuser@example.com", result.User.Email)
	assert.NotEmpty(t, result.Token)
	assert.NotEmpty(t, result.RefreshToken)
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	authSvc, userSvc := newTestAuthService(t)
	ctx := context.Background()

	createTestUser(t, userSvc, "dup@example.com", "pass")

	_, err := authSvc.Register(ctx, &RegisterRequest{
		Email:    "dup@example.com",
		Username: "other",
		Password: "pass",
	})
	assert.ErrorIs(t, err, ErrEmailExists)
}

func TestAuth_ValidateTokenExpired(t *testing.T) {
	db := testkit.SetupTestDB(t)
	userSvc := userService.NewService(infra.NewUserRepository(db))

	cfg := &Config{
		JWTExpiration:     time.Hour,
		RefreshExpiration: 24 * time.Hour,
		Issuer:            "test-issuer",
	}
	fixture := configureTestAccessTokens(t, cfg)
	authSvc := NewService(cfg, userSvc)

	token := signExpiredTestAccessToken(t, fixture, 1)
	_, err := authSvc.ValidateToken(token)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestAuth_ValidateTokenInvalid(t *testing.T) {
	authSvc, _ := newTestAuthService(t)

	_, err := authSvc.ValidateToken("not-a-jwt")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuth_ValidateTokenWrongSecret(t *testing.T) {
	authSvc, userSvc := newTestAuthService(t)
	ctx := context.Background()

	createTestUser(t, userSvc, "x@x.com", "p")
	u, _ := userSvc.GetByEmail(ctx, "x@x.com")

	pair, err := authSvc.GenerateTokenPair(u, 0, "")
	require.NoError(t, err)

	// Create a second service with a different secret
	cfg2 := &Config{
		JWTExpiration:     time.Hour,
		RefreshExpiration: 24 * time.Hour,
		Issuer:            "test-issuer",
	}
	configureTestAccessTokens(t, cfg2)
	otherSvc := NewService(cfg2, userSvc)

	_, err = otherSvc.ValidateToken(pair.AccessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuth_TokenClaimsRoundTrip(t *testing.T) {
	authSvc, userSvc := newTestAuthService(t)
	ctx := context.Background()

	createTestUser(t, userSvc, "rt@example.com", "p")
	u, _ := userSvc.GetByEmail(ctx, "rt@example.com")

	pair, err := authSvc.GenerateTokenPairWithContext(ctx, u, 7, "member")
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, u.Email, claims.Subject)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, 5*time.Second)
}
