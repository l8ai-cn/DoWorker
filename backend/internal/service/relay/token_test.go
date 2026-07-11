package relay

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGeneratePreviewToken(t *testing.T) {
	g := NewTokenGenerator("secret", "iss")
	tok, err := g.GeneratePreviewToken("pod1", 7, 42, 3, "127.0.0.1:3000", "/files/%25", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}

	parsed, err := jwt.ParseWithClaims(tok, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	claims := parsed.Claims.(*TokenClaims)
	if claims.PreviewTarget != "127.0.0.1:3000" || claims.PreviewPath != "/files/%25" {
		t.Fatalf("unexpected preview claims: %+v", claims)
	}

	if _, err := g.GeneratePreviewToken("pod1", 7, 42, 3, "", "/app", time.Hour); err == nil {
		t.Fatal("preview token without target must error")
	}
	if _, err := g.GeneratePreviewToken("pod1", 7, 42, 3, "127.0.0.1:3000", "/app/../admin", time.Hour); err == nil {
		t.Fatal("preview token with invalid path must error")
	}
}

func TestGenerateTypedToken_TunnelUnchanged(t *testing.T) {
	g := NewTokenGenerator("secret", "iss")
	tok, err := g.GenerateTypedToken("", 7, 0, 3, "tunnel", "", time.Hour)
	if err != nil || tok == "" {
		t.Fatalf("GenerateTypedToken tunnel failed: token=%q err=%v", tok, err)
	}
}

func TestNewTokenGenerator(t *testing.T) {
	g := NewTokenGenerator("secret", "issuer")
	if g == nil {
		t.Fatal("NewTokenGenerator returned nil")
	}
	if string(g.secretKey) != "secret" {
		t.Error("secret not set")
	}
	if g.issuer != "issuer" {
		t.Error("issuer not set")
	}
}

func TestGenerateToken(t *testing.T) {
	g := NewTokenGenerator("test-secret", "test-issuer")

	token, err := g.GenerateToken("pod-1", 1, 2, 3, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}

	// Parse and verify token
	parsed, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok {
		t.Fatal("invalid claims type")
	}
	if claims.PodKey != "pod-1" {
		t.Errorf("pod_key: got %q, want %q", claims.PodKey, "pod-1")
	}
	if claims.RunnerID != 1 {
		t.Errorf("runner_id: got %d, want 1", claims.RunnerID)
	}
	if claims.UserID != 2 {
		t.Errorf("user_id: got %d, want 2", claims.UserID)
	}
	if claims.OrgID != 3 {
		t.Errorf("org_id: got %d, want 3", claims.OrgID)
	}
	if claims.Issuer != "test-issuer" {
		t.Errorf("issuer: got %q, want %q", claims.Issuer, "test-issuer")
	}
	if claims.Subject != "pod-1" {
		t.Errorf("subject: got %q, want %q", claims.Subject, "pod-1")
	}
}

func TestGenerateTokenExpiry(t *testing.T) {
	g := NewTokenGenerator("test-secret", "test-issuer")

	// Short expiry - use 1 second to ensure proper testing
	token, _ := g.GenerateToken("pod-1", 1, 2, 3, 1*time.Second)

	// Should be valid now
	parsed, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Errorf("token should be valid immediately, got error: %v", err)
	}
	if parsed != nil && !parsed.Valid {
		t.Error("token should be valid immediately")
	}

	// Wait for expiry (jwt-go has some tolerance, use longer wait)
	time.Sleep(1500 * time.Millisecond)

	// Should be expired
	_, err = jwt.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err == nil {
		t.Error("token should be expired")
	}
}

func TestTokenClaimsFields(t *testing.T) {
	claims := &TokenClaims{
		PodKey:   "pod-1",
		RunnerID: 10,
		UserID:   20,
		OrgID:    30,
	}

	if claims.PodKey != "pod-1" {
		t.Error("PodKey")
	}
	if claims.RunnerID != 10 {
		t.Error("RunnerID")
	}
	if claims.UserID != 20 {
		t.Error("UserID")
	}
	if claims.OrgID != 30 {
		t.Error("OrgID")
	}
}
