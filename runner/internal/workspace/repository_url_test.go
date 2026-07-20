package workspace

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRepositoryURLRejectsHTTPUserinfo(t *testing.T) {
	for _, raw := range []string{
		"https://token@example.test/org/repo.git",
		"https://user:secret@example.test/org/repo.git",
		"http://user@example.test/org/repo.git",
		"HTTPS://user:secret@example.test/org/repo.git",
	} {
		t.Run(raw, func(t *testing.T) {
			err := validateRepositoryURL(raw)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "must not contain userinfo")
		})
	}
}

func TestValidateRepositoryURLParseErrorDoesNotLeakCredentials(t *testing.T) {
	err := validateRepositoryURL("https://user:secret%zz@example.test/org/repo.git")

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "user")
	assert.NotContains(t, err.Error(), "secret")
}

func TestValidateRepositoryURLRejectsQueryAndFragmentWithoutLeakingValues(t *testing.T) {
	for _, raw := range []string{
		"https://example.test/org/repo.git?access_token=query-secret",
		"https://example.test/org/repo.git#fragment-secret",
	} {
		err := validateRepositoryURL(raw)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not contain query or fragment")
		assert.NotContains(t, err.Error(), "secret")
	}
}

func TestValidateRepositoryURLAcceptsSupportedRepositoryForms(t *testing.T) {
	for _, raw := range []string{
		"https://example.test/org/repo.git",
		"http://example.test/org/repo.git",
		"ssh://git@example.test/org/repo.git",
		"git@example.test:org/repo.git",
		"/tmp/repo.git",
	} {
		t.Run(raw, func(t *testing.T) {
			require.NoError(t, validateRepositoryURL(raw))
		})
	}
}

func TestRepositoryURLForDisplayStripsCredentialsAndQuery(t *testing.T) {
	display := RepositoryURLForDisplay("HTTPS://user:secret@example.test/org/repo.git?token=hidden#fragment")

	assert.NotContains(t, display, "user")
	assert.NotContains(t, display, "secret")
	assert.NotContains(t, display, "hidden")
	assert.Contains(t, strings.ToLower(display), "https://example.test/org/repo.git")
}
