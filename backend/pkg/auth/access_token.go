package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrAccessTokenConfig = errors.New("invalid access token configuration")

type AccessTokenConfig struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KeyID      string
	Issuer     string
	Audiences  []string
	Duration   time.Duration
}

type AccessTokenManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	issuer     string
	audiences  []string
	duration   time.Duration
}

type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

func NewAccessTokenManager(config AccessTokenConfig) (*AccessTokenManager, error) {
	if config.PublicKey == nil ||
		config.PublicKey.N.BitLen() < 2048 ||
		strings.TrimSpace(config.KeyID) == "" ||
		strings.TrimSpace(config.Issuer) == "" ||
		len(config.Audiences) == 0 ||
		config.Duration <= 0 {
		return nil, ErrAccessTokenConfig
	}
	for _, audience := range config.Audiences {
		if strings.TrimSpace(audience) == "" {
			return nil, ErrAccessTokenConfig
		}
	}
	if config.PrivateKey != nil &&
		(config.PrivateKey.PublicKey.N.Cmp(config.PublicKey.N) != 0 ||
			config.PrivateKey.PublicKey.E != config.PublicKey.E) {
		return nil, ErrAccessTokenConfig
	}
	return &AccessTokenManager{
		privateKey: config.PrivateKey,
		publicKey:  config.PublicKey,
		keyID:      strings.TrimSpace(config.KeyID),
		issuer:     strings.TrimSpace(config.Issuer),
		audiences:  append([]string(nil), config.Audiences...),
		duration:   config.Duration,
	}, nil
}

func (m *AccessTokenManager) GenerateToken(
	userID int64,
	email string,
	username string,
	orgID int64,
	role string,
) (string, error) {
	if m.privateKey == nil {
		return "", ErrAccessTokenConfig
	}
	now := time.Now()
	claims := Claims{
		UserID:         userID,
		Email:          email,
		Username:       username,
		OrganizationID: orgID,
		Role:           role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   email,
			Audience:  jwt.ClaimStrings(m.audiences),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = m.keyID
	return token.SignedString(m.privateKey)
}

func (m *AccessTokenManager) ValidateToken(
	tokenString string,
	requiredAudience string,
) (*Claims, error) {
	if !slices.Contains(m.audiences, requiredAudience) {
		return nil, ErrInvalidToken
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodRS256 || token.Header["kid"] != m.keyID {
				return nil, ErrInvalidToken
			}
			return m.publicKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(requiredAudience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *AccessTokenManager) JWKS() JSONWebKeySet {
	return JSONWebKeySet{Keys: []JSONWebKey{{
		KeyType:   "RSA",
		Use:       "sig",
		Algorithm: jwt.SigningMethodRS256.Alg(),
		KeyID:     m.keyID,
		Modulus:   base64.RawURLEncoding.EncodeToString(m.publicKey.N.Bytes()),
		Exponent:  base64.RawURLEncoding.EncodeToString(big.NewInt(int64(m.publicKey.E)).Bytes()),
	}}}
}
