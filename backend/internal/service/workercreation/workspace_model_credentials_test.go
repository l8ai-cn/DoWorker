package workercreation

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
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

func TestWorkspaceResolverAllowsCredentialBundleFieldInRuntimeBundle(t *testing.T) {
	fixture := newWorkspaceFixture()
	fixture.envBundles.rows[5].Data = envbundle.BundleData{
		"CURSOR_API_KEY": "must-not-be-treated-as-a-model-resource",
	}
	deps := fixture.deps()
	deps.Definitions = staticWorkerDefinitions{
		"cursor-cli": {
			Slug: "cursor-cli",
			ModelRequirement: workerdefinition.ModelRequirement{
				Required: false,
			},
			CredentialBindings: []workerdefinition.CredentialBinding{{
				ID: "cursor",
				Source: workerdefinition.CredentialSource{
					Kind: "credential_bundle",
					Ref:  "cursor",
				},
				Target: workerdefinition.CredentialTarget{
					Kind: "env",
					Name: "CURSOR_API_KEY",
				},
			}},
		},
	}

	_, err := newWorkspaceResolver(deps).ResolveWorkspace(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("cursor-cli"),
		validWorkspaceDraft(),
	)

	require.NoError(t, err)
}
