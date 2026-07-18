package workercreation

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceResolverAllowsOmittedOptionalConfigDocument(t *testing.T) {
	resolver := configDocumentResolver(t)

	documentIDs, err := resolver.resolveConfigDocumentIDs(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
		[]specdomain.ConfigDocumentBinding{{
			DocumentID:     "settings",
			ConfigBundleID: 7,
		}},
	)

	require.NoError(t, err)
	assert.Equal(t, []string{"settings"}, documentIDs)
}

func TestWorkspaceResolverRejectsOmittedRequiredConfigDocument(t *testing.T) {
	resolver := configDocumentResolver(t)

	_, err := resolver.resolveConfigDocumentIDs(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
		nil,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "settings")
}

func TestCompilerUsesConfigDocumentID(t *testing.T) {
	spec := validCompiledWorkerSpec()
	spec.Workspace.ConfigDocumentBindings = []specdomain.ConfigDocumentBinding{{
		DocumentID:     "settings",
		ConfigBundleID: 7,
	}}

	layer, err := newCompiler(configDocumentResolver(t)).Compile(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		spec,
	)

	require.NoError(t, err)
	assert.Contains(t, layer, "USE_CONFIG_BUNDLE settings\n")
}

func configDocumentResolver(t *testing.T) *workspaceResolver {
	t.Helper()
	fixture := newWorkspaceFixture()
	fixture.envBundles.rows[7] = &envbundle.EnvBundle{
		ID:         7,
		OwnerScope: envbundle.OwnerScopeUser,
		OwnerID:    7,
		Name:       "settings-document",
		Kind:       envbundle.KindConfig,
		IsActive:   true,
	}
	definition, ok := fixture.deps().Definitions.Get("codex-cli")
	require.True(t, ok)
	definition.ConfigDocuments = []workerdefinition.ConfigDocument{
		{
			ID:         "settings",
			Format:     "json",
			TargetPath: "/workspace/settings.json",
			Required:   true,
		},
		{
			ID:         "optional",
			Format:     "json",
			TargetPath: "/workspace/optional.json",
			Required:   false,
		},
	}
	deps := fixture.deps()
	deps.Definitions = staticWorkerDefinitions{"codex-cli": definition}
	return newWorkspaceResolver(deps)
}
