package workercreation

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceResolverRejectsModelCredentialsInRuntimeBundle(t *testing.T) {
	fixture := newWorkspaceFixture()
	fixture.envBundles.rows[5].Data = envbundle.BundleData{
		"OPENAI_API_KEY": "must-come-from-model-resource",
	}

	_, err := newWorkspaceResolver(fixture.deps()).ResolveWorkspace(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
		validWorkspaceDraft(),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "model resource")
}
