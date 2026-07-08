package v1

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func TestValidateAvatarInput_NilOptional(t *testing.T) {
	out, err := validateAvatarInput(nil)
	require.NoError(t, err)
	assert.Nil(t, out)

	out, err = validateAvatarInput(&avatarInput{ContentBase64: "  "})
	require.NoError(t, err)
	assert.Nil(t, out)
}

func TestValidateAvatarInput_SniffsAllowedTypes(t *testing.T) {
	png := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 32)...)
	out, err := validateAvatarInput(&avatarInput{Filename: "evil.exe", ContentBase64: b64(png)})
	require.NoError(t, err)
	require.NotNil(t, out)
	// Client filename ignored; platform derives the extension from magic bytes.
	assert.Equal(t, "png", out.Ext)

	gif := append([]byte("GIF89a"), make([]byte, 16)...)
	out, err = validateAvatarInput(&avatarInput{ContentBase64: b64(gif)})
	require.NoError(t, err)
	assert.Equal(t, "gif", out.Ext)
}

func TestValidateAvatarInput_RejectsBadBase64(t *testing.T) {
	_, err := validateAvatarInput(&avatarInput{ContentBase64: "!!!not base64!!!"})
	assert.Error(t, err)
}

func TestValidateAvatarInput_RejectsDisallowedType(t *testing.T) {
	// A plain-text / octet blob sniffs to something outside the allow-list.
	_, err := validateAvatarInput(&avatarInput{ContentBase64: b64([]byte("just some text, not an image at all"))})
	assert.Error(t, err)
}

func TestValidateAvatarInput_RejectsOversize(t *testing.T) {
	big := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, maxAvatarBytes+1)...)
	_, err := validateAvatarInput(&avatarInput{ContentBase64: b64(big)})
	assert.Error(t, err)
}

func TestSanitizeRepoPath_Accepts(t *testing.T) {
	cases := map[string]string{
		"/agent.md":           "agent.md",
		"agent.md":            "agent.md",
		"/assets/avatar.png":  "assets/avatar.png",
		"assets/./avatar.png": "assets/avatar.png",
		"/wiki/sub/page.md":   "wiki/sub/page.md",
	}
	for in, want := range cases {
		got, err := sanitizeRepoPath(in)
		require.NoErrorf(t, err, "input %q", in)
		assert.Equalf(t, want, got, "input %q", in)
	}
}

func TestSanitizeRepoPath_Rejects(t *testing.T) {
	bad := []string{
		"",
		"/",
		"..",
		"../secret",
		"/../secret",
		"assets/../../etc/passwd",
		"a/../../b",
		"foo\x00bar",
		"line\nbreak",
	}
	for _, in := range bad {
		_, err := sanitizeRepoPath(in)
		assert.Errorf(t, err, "expected rejection for %q", in)
	}
}
