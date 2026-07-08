package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPrimaryCredentialResolver struct {
	name string
}

func (s stubPrimaryCredentialResolver) PrimaryCredentialBundleName(context.Context, int64, int64, string) (string, error) {
	return s.name, nil
}

func TestAppendPrimaryCredentialBundle_AppendsWhenMissing(t *testing.T) {
	var layer *string
	AppendPrimaryCredentialBundle(context.Background(), stubPrimaryCredentialResolver{name: "codex"}, 1, 2, "codex-cli", &layer)
	require.NotNil(t, layer)
	assert.Contains(t, *layer, `USE_ENV_BUNDLE "codex"`)
}

func TestAppendPrimaryCredentialBundle_SkipsDuplicate(t *testing.T) {
	existing := `MODE acp` + "\n" + `USE_ENV_BUNDLE "codex"`
	layer := &existing
	AppendPrimaryCredentialBundle(context.Background(), stubPrimaryCredentialResolver{name: "codex"}, 1, 2, "codex-cli", &layer)
	assert.Equal(t, existing, *layer)
}

func TestAppendPrimaryCredentialBundle_NoResolver(t *testing.T) {
	var layer *string
	AppendPrimaryCredentialBundle(context.Background(), nil, 1, 2, "codex-cli", &layer)
	assert.Nil(t, layer)
}
