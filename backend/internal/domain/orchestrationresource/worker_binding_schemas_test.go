package orchestrationresource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterWorkerSchemasRegistersEveryWorkerResourceKind(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))

	kinds := []string{
		KindModelBinding,
		KindRepository,
		KindSkill,
		KindKnowledgeBase,
		KindEnvironmentBundle,
		KindComputeTarget,
		KindResourceProfile,
		KindToolBinding,
		KindWorkerTemplate,
	}
	for _, kind := range kinds {
		require.True(t, registry.Has(TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       kind,
		}), kind)
	}
}

func TestRegisterWorkerSchemasReturnsDuplicateRegistrationError(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))

	err := RegisterWorkerSchemas(registry)
	require.ErrorIs(t, err, ErrDuplicateSchema)
	require.NotContains(t, err.Error(), "register worker schema")
}

func TestIDBindingSchemasRequirePositiveIDs(t *testing.T) {
	tests := []struct {
		kind  string
		field string
		spec  any
	}{
		{KindModelBinding, "resourceId", ModelBindingSpec{ResourceID: 0}},
		{KindRepository, "repositoryId", RepositoryBindingSpec{RepositoryID: 0}},
		{KindSkill, "skillId", SkillBindingSpec{SkillID: 0}},
		{KindKnowledgeBase, "knowledgeBaseId", KnowledgeBaseBindingSpec{KnowledgeBaseID: 0}},
		{KindEnvironmentBundle, "environmentBundleId", EnvironmentBundleBindingSpec{EnvironmentBundleID: 0}},
		{KindComputeTarget, "computeTargetId", ComputeTargetBindingSpec{ComputeTargetID: 0}},
		{KindResourceProfile, "resourceProfileId", ResourceProfileBindingSpec{ResourceProfileID: 0}},
	}
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))

	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			_, err := registry.DecodeAndValidate(
				workerSchemaManifest(t, test.kind, test.spec),
			)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.field)
		})
	}
}

func TestToolBindingRequiresModelBindingReference(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))

	valid := ToolBindingSpec{
		ModelRef: workerDraftReference(KindModelBinding, "coding-primary"),
	}
	decoded, err := registry.DecodeAndValidate(
		workerSchemaManifest(t, KindToolBinding, valid),
	)
	require.NoError(t, err)
	require.Equal(t, &valid, decoded)

	valid.ModelRef.Kind = KindRepository
	_, err = registry.DecodeAndValidate(
		workerSchemaManifest(t, KindToolBinding, valid),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "modelRef")
	require.Contains(t, err.Error(), KindModelBinding)
}
