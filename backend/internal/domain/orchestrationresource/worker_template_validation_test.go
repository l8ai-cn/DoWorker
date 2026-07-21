package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/require"
)

func TestWorkerTemplateRejectsInvalidOptionsRevision(t *testing.T) {
	tests := []string{
		"",
		" catalog-1",
		"catalog-1 ",
		"catalog\n1",
		"catalog-\u202e1",
		strings.Repeat("a", 129),
	}
	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			spec := validWorkerTemplateSpec()
			spec.OptionsRevision = value
			requireWorkerTemplateError(t, spec, "optionsRevision")
		})
	}
}

func TestWorkerTemplateRejectsInvalidMapKeys(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		mutate func(*WorkerTemplateSpec)
	}{
		{
			name:  "tool role",
			input: "Video_Generation",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.ToolRefs["Video_Generation"] = spec.ToolRefs["video-generation"]
				delete(spec.ToolRefs, "video-generation")
			},
		},
		{
			name:  "config value field",
			input: " invalid ",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.Values[" invalid "] = "high"
				delete(spec.TypeConfig.Values, "reasoning-effort")
			},
		},
		{
			name:  "secret config field",
			input: " invalid ",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.SecretRefs[" invalid "] = spec.TypeConfig.SecretRefs["api-token"]
				delete(spec.TypeConfig.SecretRefs, "api-token")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validWorkerTemplateSpec()
			test.mutate(&spec)
			err := workerTemplateError(t, spec)
			require.Contains(t, err.Error(), test.name)
			require.NotContains(t, err.Error(), test.input)
		})
	}
}

func TestWorkerTemplateAcceptsDefinitionConfigFieldNames(t *testing.T) {
	spec := validWorkerTemplateSpec()
	spec.TypeConfig.Values["REASONING_EFFORT"] = "high"
	delete(spec.TypeConfig.Values, "reasoning-effort")
	spec.TypeConfig.SecretRefs["CURSOR_API_KEY"] =
		spec.TypeConfig.SecretRefs["api-token"]
	delete(spec.TypeConfig.SecretRefs, "api-token")

	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))
	_, err := registry.DecodeAndValidate(
		workerSchemaManifest(t, KindWorkerTemplate, spec),
	)
	require.NoError(t, err)
}

func TestWorkerTemplateRejectsWrongKindCrossNamespaceAndResolvedFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*WorkerTemplateSpec)
		field  string
	}{
		{
			name: "wrong model kind",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.ModelRef.Kind = KindRepository
			},
			field: "modelRef",
		},
		{
			name: "cross namespace",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.SkillRefs[0].Namespace = "team-beta"
			},
			field: "workspace.skillRefs",
		},
		{
			name: "uid",
			mutate: func(spec *WorkerTemplateSpec) {
				ref := spec.ToolRefs["video-generation"]
				ref.UID = "resolved-uid"
				spec.ToolRefs["video-generation"] = ref
			},
			field: "reference.uid",
		},
		{
			name: "digest",
			mutate: func(spec *WorkerTemplateSpec) {
				ref := spec.TypeConfig.SecretRefs["api-token"]
				ref.Digest = "sha256:" + strings.Repeat("a", 64)
				spec.TypeConfig.SecretRefs["api-token"] = ref
			},
			field: "reference.digest",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validWorkerTemplateSpec()
			test.mutate(&spec)
			requireWorkerTemplateError(t, spec, test.field)
		})
	}
}

func TestWorkerTemplateRejectsResourceProfileAndCustomResourcesTogether(t *testing.T) {
	spec := validWorkerTemplateSpec()
	spec.Runtime.CustomResources = &workerspec.ResourceRequestsLimits{}
	requireWorkerTemplateError(t, spec, "resourceProfileRef")
}

func TestWorkerTemplateRejectsDuplicateReferencesInCollections(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*WorkerTemplateSpec)
		field  string
	}{
		{
			name: "tool refs",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.ToolRefs["image-generation"] = spec.ToolRefs["video-generation"]
			},
			field: "toolRefs",
		},
		{
			name: "skills",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.SkillRefs = append(
					spec.Workspace.SkillRefs,
					spec.Workspace.SkillRefs[0],
				)
			},
			field: "workspace.skillRefs",
		},
		{
			name: "knowledge",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.KnowledgeMounts = append(
					spec.Workspace.KnowledgeMounts,
					spec.Workspace.KnowledgeMounts[0],
				)
			},
			field: "workspace.knowledgeMounts",
		},
		{
			name: "environment bundles",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.EnvironmentBundleRefs = append(
					spec.Workspace.EnvironmentBundleRefs,
					spec.Workspace.EnvironmentBundleRefs[0],
				)
			},
			field: "workspace.environmentBundleRefs",
		},
		{
			name: "config document bindings",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.ConfigDocumentBindings = append(
					spec.Workspace.ConfigDocumentBindings,
					spec.Workspace.ConfigDocumentBindings[0],
				)
			},
			field: "workspace.configDocumentBindings",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validWorkerTemplateSpec()
			test.mutate(&spec)
			requireWorkerTemplateError(t, spec, test.field)
		})
	}
}

func TestWorkerTemplateAllowsOneCredentialBundleForMultipleSecretFields(t *testing.T) {
	spec := validWorkerTemplateSpec()
	spec.TypeConfig.SecretRefs["access-token"] = spec.TypeConfig.SecretRefs["api-token"]

	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))
	_, err := registry.DecodeAndValidate(
		workerSchemaManifest(t, KindWorkerTemplate, spec),
	)
	require.NoError(t, err)
}

func requireWorkerTemplateError(
	t *testing.T,
	spec WorkerTemplateSpec,
	message string,
) {
	t.Helper()
	err := workerTemplateError(t, spec)
	require.Contains(t, err.Error(), message)
}

func workerTemplateError(t *testing.T, spec WorkerTemplateSpec) error {
	t.Helper()
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))
	_, err := registry.DecodeAndValidate(
		workerSchemaManifest(t, KindWorkerTemplate, spec),
	)
	require.Error(t, err)
	return err
}
