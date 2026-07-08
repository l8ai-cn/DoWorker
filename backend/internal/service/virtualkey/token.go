package virtualkey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const tokenPrefix = "dwk_"

// generatedToken carries the one-time plaintext plus the persisted derivations.
type generatedToken struct {
	Token  string // plaintext, shown once
	Prefix string // display prefix (dwk_ + first 8 chars)
	Hash   string // sha256 hex, stored for lookup
}

func newToken() (generatedToken, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return generatedToken{}, err
	}
	body := hex.EncodeToString(buf)
	token := tokenPrefix + body
	return generatedToken{
		Token:  token,
		Prefix: token[:12],
		Hash:   hashToken(token),
	}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
