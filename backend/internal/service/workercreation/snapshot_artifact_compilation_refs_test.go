package workercreation

import (
	"testing"

	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtifactCompilationReferencesDeduplicateSharedSecretBundle(t *testing.T) {
	resolver := newArtifactCompilationResolver(workerdependency.Document{
		SecretReferences: []workerdependency.SecretReference{
			secretArtifactReference("ACCESS_KEY", "lovart", 6),
			secretArtifactReference("SECRET_KEY", "lovart", 6),
		},
	})

	refs, err := resolver.compilationReferences(map[string]specdomain.SecretReference{
		"ACCESS_KEY": {Kind: slugkit.MustNewForTest("env-bundle"), ID: 6},
		"SECRET_KEY": {Kind: slugkit.MustNewForTest("env-bundle"), ID: 6},
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"lovart"}, refs.EnvBundleNames)
}

func secretArtifactReference(
	field string,
	name string,
	domainID int64,
) workerdependency.SecretReference {
	return workerdependency.SecretReference{
		Pin: workerdependency.ResourcePin{
			DomainID: domainID,
			Reference: resource.Reference{
				Kind: resource.KindEnvironmentBundle,
				Name: slugkit.MustNewForTest(name),
			},
		},
		Field: field,
	}
}
