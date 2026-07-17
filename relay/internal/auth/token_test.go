package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret = "test-secret-key-for-testing"
	testIssuer = "test-issuer"
)

func TestNewTokenValidator(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	if v == nil || string(v.secretKey) != testSecret || v.issuer != testIssuer {
		t.Error("NewTokenValidator failed")
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(testSecret, testIssuer, "pod-1", 1, 2, 3, time.Hour)
	if err != nil || token == "" {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	v := NewTokenValidator(testSecret, testIssuer)
	claims, err := v.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.PodKey != "pod-1" ||
		claims.RunnerID != 1 || claims.UserID != 2 || claims.OrgID != 3 ||
		claims.Issuer != testIssuer || claims.Subject != "pod-1" {
		t.Error("claims mismatch")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	token, _ := GenerateToken(testSecret, testIssuer, "pod-1", 1, 2, 3, time.Hour)
	if claims, err := v.ValidateToken(token); err != nil || claims == nil {
		t.Errorf("ValidateToken failed: %v", err)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	token, _ := GenerateToken(testSecret, testIssuer, "pod-1", 1, 2, 3, -time.Hour)
	if _, err := v.ValidateToken(token); err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	token, _ := GenerateToken("wrong-secret", testIssuer, "pod-1", 1, 2, 3, time.Hour)
	if _, err := v.ValidateToken(token); err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_InvalidIssuer(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	token, _ := GenerateToken(testSecret, "wrong-issuer", "pod-1", 1, 2, 3, time.Hour)
	if _, err := v.ValidateToken(token); err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_NoIssuerCheck(t *testing.T) {
	v := NewTokenValidator(testSecret, "")
	token, _ := GenerateToken(testSecret, "any-issuer", "pod-1", 1, 2, 3, time.Hour)
	claims, err := v.ValidateToken(token)
	if err != nil || claims.Issuer != "any-issuer" {
		t.Error("should succeed when issuer check disabled")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	for _, token := range []string{"", "not-a-token", "eyJhbGciOiJIUzI1NiJ9", "eyJ.!!!.xyz"} {
		if _, err := v.ValidateToken(token); err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken for %q", token)
		}
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	claims := &RelayClaims{PodKey: "pod-1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()), Issuer: testIssuer}}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := v.ValidateToken(tokenString); err != ErrInvalidToken {
		t.Error("expected ErrInvalidToken for wrong signing method")
	}
}

func TestValidateToken_RejectsMissingExpiration(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	claims := &RelayClaims{
		PodKey: "pod-1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: testIssuer,
		},
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.ValidateToken(tokenString); err != ErrInvalidToken {
		t.Fatalf("missing expiration error = %v, want ErrInvalidToken", err)
	}
}

func TestValidateToken_RejectsNonHS256HMAC(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	for _, method := range []jwt.SigningMethod{jwt.SigningMethodHS384, jwt.SigningMethodHS512} {
		claims := &RelayClaims{
			PodKey: "pod-1",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				Issuer:    testIssuer,
			},
		}
		tokenString, err := jwt.NewWithClaims(method, claims).SignedString([]byte(testSecret))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := v.ValidateToken(tokenString); err != ErrInvalidToken {
			t.Fatalf("%s error = %v, want ErrInvalidToken", method.Alg(), err)
		}
	}
}

func TestRelayClaims_AllFields(t *testing.T) {
	now := time.Now()
	claims := &RelayClaims{PodKey: "p1", RunnerID: 100, UserID: 200, OrgID: 300,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)), IssuedAt: jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now), Issuer: testIssuer, Subject: "p1"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))
	v := NewTokenValidator(testSecret, testIssuer)
	decoded, err := v.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if decoded.PodKey != "p1" ||
		decoded.RunnerID != 100 || decoded.UserID != 200 || decoded.OrgID != 300 ||
		decoded.Issuer != testIssuer || decoded.Subject != "p1" {
		t.Error("decoded claims mismatch")
	}
}

func TestRelayClaims_TokenType(t *testing.T) {
	secret := "s3cret"
	// 旧 token（无 token_type）：runner=UserID 0，browser=UserID!=0 仍成立
	legacyRunner, err := GenerateToken(secret, "iss", "pod1", 7, 0, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewTokenValidator(secret, "iss")
	c, err := v.ValidateToken(legacyRunner)
	if err != nil {
		t.Fatal(err)
	}
	if !c.IsRunnerToken() {
		t.Fatalf("legacy runner token should be runner")
	}
	if c.ResolvedType() != TokenTypeRunner {
		t.Fatalf("legacy runner should resolve to runner, got %q", c.ResolvedType())
	}

	// 新 token：显式 tunnel 类型
	tunnel, err := GenerateTypedToken(secret, "iss", TokenTypeTunnel, "", 7, 0, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	tc, err := v.ValidateToken(tunnel)
	if err != nil {
		t.Fatal(err)
	}
	if tc.ResolvedType() != TokenTypeTunnel {
		t.Fatalf("expected tunnel, got %q", tc.ResolvedType())
	}
}

func TestValidatePreviewTokenRequiresBoundClaims(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	for _, previewPath := range []string{"", "/app/", "/app/../admin", "/app%2F..%2Fadmin", "/files/%"} {
		previewPath := previewPath
		t.Run(previewPath, func(t *testing.T) {
			token := signPreviewClaims(t, TokenTypePreviewBootstrap, previewPath, "https://preview.example.com")
			if _, err := v.ValidatePreviewToken(token, TokenTypePreviewBootstrap, "https://preview.example.com"); err != ErrInvalidToken {
				t.Fatalf("ValidateToken path %q error = %v, want ErrInvalidToken", previewPath, err)
			}
		})
	}

	for _, previewPath := range []string{
		"/app",
		"/files/%25",
		"/files/report%23draft.pdf",
		"/route/%3F",
		"/app/%252e%252e",
	} {
		claims, err := v.ValidatePreviewToken(
			signPreviewClaims(t, TokenTypePreviewBootstrap, previewPath, "https://preview.example.com"),
			TokenTypePreviewBootstrap,
			"https://preview.example.com",
		)
		if err != nil {
			t.Fatalf("normalized preview path %q rejected: %v", previewPath, err)
		}
		if claims.PreviewPath != previewPath {
			t.Fatalf("PreviewPath = %q, want %q", claims.PreviewPath, previewPath)
		}
	}
}

func TestValidatePreviewTokenRejectsWrongOriginAndType(t *testing.T) {
	v := NewTokenValidator(testSecret, testIssuer)
	token := signPreviewClaims(t, TokenTypePreviewBootstrap, "/app", "https://preview.example.com")
	if _, err := v.ValidatePreviewToken(token, TokenTypePreviewBootstrap, "https://other.example.com"); err != ErrInvalidToken {
		t.Fatalf("wrong origin error = %v", err)
	}
	if _, err := v.ValidatePreviewToken(token, TokenTypePreviewSession, "https://preview.example.com"); err != ErrInvalidToken {
		t.Fatalf("wrong type error = %v", err)
	}
}

func TestNormalizePreviewPath_Idempotent(t *testing.T) {
	for _, previewPath := range []string{
		"/files/%25",
		"/files/report%23draft.pdf",
		"/route/%3F",
		"/app/%252e%252e",
		"/documents/%E4%B8%AD",
	} {
		normalized, err := NormalizePreviewPath(previewPath)
		if err != nil {
			t.Fatalf("NormalizePreviewPath(%q): %v", previewPath, err)
		}
		again, err := NormalizePreviewPath(normalized)
		if err != nil || again != normalized {
			t.Fatalf("second normalization: first=%q second=%q err=%v", normalized, again, err)
		}
	}
}

func signPreviewClaims(t *testing.T, tokenType TokenType, previewPath, previewOrigin string) string {
	t.Helper()
	now := time.Now()
	claims := &RelayClaims{
		PodKey:        "pod1",
		RunnerID:      7,
		UserID:        42,
		OrgID:         3,
		TokenType:     tokenType,
		PreviewTarget: "127.0.0.1:3000",
		PreviewPath:   previewPath,
		PreviewOrigin: previewOrigin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    testIssuer,
			Subject:   "pod1",
			ID:        "jti-1",
			Audience:  jwt.ClaimStrings{previewOrigin},
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestErrorVariables(t *testing.T) {
	if ErrInvalidToken.Error() != "invalid token" {
		t.Error("ErrInvalidToken message wrong")
	}
	if ErrTokenExpired.Error() != "token expired" {
		t.Error("ErrTokenExpired message wrong")
	}
}
