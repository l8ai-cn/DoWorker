package orchestrationresource

import (
	"encoding/json"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/require"
)

func TestWorkerTemplateSchemaDecodesCompleteCamelCaseDraft(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))
	expected := validWorkerTemplateSpec()

	decoded, err := registry.DecodeAndValidate(
		workerSchemaManifest(t, KindWorkerTemplate, expected),
	)
	require.NoError(t, err)
	require.Equal(t, &expected, decoded)

	canonical, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.Contains(t, string(canonical), `"optionsRevision"`)
	require.Contains(t, string(canonical), `"runtimeImageId"`)
	require.Contains(t, string(canonical), `"environmentBundleRefs"`)
	require.NotContains(t, string(canonical), `"options_revision"`)
	require.NotContains(t, string(canonical), `"runtime_image_id"`)
}

func TestWorkerTemplateCustomResourcesUseCamelCaseFields(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))
	spec := validWorkerTemplateSpec()
	spec.Runtime.ResourceProfileRef = nil
	gpu := uint32(1)
	spec.Runtime.CustomResources = &workerspec.ResourceRequestsLimits{
		CPURequestMilliCPU:  500,
		CPULimitMilliCPU:    1_000,
		MemoryRequestBytes:  1 << 30,
		MemoryLimitBytes:    2 << 30,
		StorageRequestBytes: 10 << 30,
		StorageLimitBytes:   20 << 30,
		GPURequest:          &gpu,
		GPULimit:            &gpu,
	}

	raw, err := json.Marshal(spec)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"cpuRequestMilliCPU"`)
	require.Contains(t, string(raw), `"storageLimitBytes"`)
	require.NotContains(t, string(raw), `"cpu_request_millicpu"`)

	decoded, err := registry.DecodeAndValidate(workerSchemaManifest(
		t,
		KindWorkerTemplate,
		json.RawMessage(raw),
	))
	require.NoError(t, err)
	require.Equal(t, &spec, decoded)
}

func TestWorkerTemplateSchemaRoundTripsCamelCaseYAML(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))
	expected := validWorkerTemplateSpec()
	manifest := workerSchemaManifest(t, KindWorkerTemplate, expected)

	encoded, err := EncodeYAML(manifest)
	require.NoError(t, err)
	require.Contains(t, string(encoded), "optionsRevision:")
	require.Contains(t, string(encoded), "runtimeImageId:")
	require.Contains(t, string(encoded), "environmentBundleRefs:")

	decodedManifest, err := DecodeYAMLSubmission(encoded)
	require.NoError(t, err)
	decoded, err := registry.DecodeAndValidate(decodedManifest)
	require.NoError(t, err)
	require.Equal(t, &expected, decoded)
}

func TestWorkerTemplateRegistryStillRejectsUnknownFields(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterWorkerSchemas(registry))
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "runtime field",
			mutate: func(object map[string]any) {
				runtime := object["runtime"].(map[string]any)
				runtime["unexpectedRuntimeField"] = true
			},
		},
		{
			name: "initial task",
			mutate: func(object map[string]any) {
				workspace := object["workspace"].(map[string]any)
				workspace["initialTask"] = "must belong to an invocation"
			},
		},
		{
			name: "source expert id",
			mutate: func(object map[string]any) {
				metadata := object["metadata"].(map[string]any)
				metadata["sourceExpertId"] = 42
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := workerSchemaManifest(
				t,
				KindWorkerTemplate,
				validWorkerTemplateSpec(),
			)
			var object map[string]any
			require.NoError(t, json.Unmarshal(manifest.Spec, &object))
			test.mutate(object)
			manifest.Spec, _ = json.Marshal(object)

			_, err := registry.DecodeAndValidate(manifest)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrTypedJSONUnknownField)
		})
	}
}
