package virtualkey

import (
	"strings"
	"testing"
)

func TestNewTokenFormatAndHash(t *testing.T) {
	tok, err := newToken()
	if err != nil {
		t.Fatalf("newToken: %v", err)
	}
	if !strings.HasPrefix(tok.Token, tokenPrefix) {
		t.Fatalf("token %q missing prefix %q", tok.Token, tokenPrefix)
	}
	if len(tok.Prefix) != 12 || !strings.HasPrefix(tok.Token, tok.Prefix) {
		t.Fatalf("prefix %q not a 12-char lead of token %q", tok.Prefix, tok.Token)
	}
	if len(tok.Hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(tok.Hash))
	}
	if got := hashToken(tok.Token); got != tok.Hash {
		t.Fatalf("hashToken mismatch: %q vs %q", got, tok.Hash)
	}
}

func TestNewTokenUnique(t *testing.T) {
	a, _ := newToken()
	b, _ := newToken()
	if a.Token == b.Token || a.Hash == b.Hash {
		t.Fatalf("expected distinct tokens, got %q and %q", a.Token, b.Token)
	}
}
