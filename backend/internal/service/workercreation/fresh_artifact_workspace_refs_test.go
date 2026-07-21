package workercreation

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendFreshWorkspaceReferencesDeduplicatesSharedSecretBundle(t *testing.T) {
	agentSlug := "pattern-designer"
	workspace := &workspaceResolver{
		envBundles: map[int64]*envbundle.EnvBundle{
			6: {
				ID: 6, OwnerScope: envbundle.OwnerScopeUser, OwnerID: 1,
				AgentSlug: &agentSlug, Name: "lovart",
				Kind: envbundle.KindCredential, IsActive: true,
			},
		},
	}
	refs := ArtifactReferences{}
	spec := specdomain.Spec{
		TypeConfig: specdomain.TypeConfig{
			SecretRefs: map[string]specdomain.SecretReference{
				"ACCESS_KEY": {Kind: slugkit.MustNewForTest("env-bundle"), ID: 6},
				"SECRET_KEY": {Kind: slugkit.MustNewForTest("env-bundle"), ID: 6},
			},
		},
	}

	err := appendFreshWorkspaceReferences(
		specservice.Scope{OrgID: 1, UserID: 1},
		slugkit.MustNewForTest("dev-org"),
		&refs,
		spec,
		workspace,
	)

	require.NoError(t, err)
	assert.Len(t, refs.SecretBundles, 2)
	assert.Len(t, refs.AllPlanReferences, 1)
}
