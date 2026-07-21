package orchestrationresource

import (
	"encoding/json"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
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

func TestWorkerTemplateRuntimeRejectsExplicitNulls(t *testing.T) {
	tests := []struct {
		path   string
		mutate func(map[string]any)
	}{
		{
			path: "spec.runtime.runtimeImageId",
			mutate: func(runtime map[string]any) {
				runtime["runtimeImageId"] = nil
			},
		},
		{
			path: "spec.runtime.resourceProfileRef",
			mutate: func(runtime map[string]any) {
				runtime["resourceProfileRef"] = nil
			},
		},
		{
			path: "spec.runtime.computeTargetRef.apiVersion",
			mutate: func(runtime map[string]any) {
				reference := runtime["computeTargetRef"].(map[string]any)
				reference["apiVersion"] = nil
			},
		},
		{
			path: "spec.runtime.computeTargetRef.revision",
			mutate: func(runtime map[string]any) {
				reference := runtime["computeTargetRef"].(map[string]any)
				reference["revision"] = nil
			},
		},
		{
			path: "spec.runtime.resourceProfileRef.kind",
			mutate: func(runtime map[string]any) {
				runtime["resourceProfileRef"] = map[string]any{
					"kind": nil,
					"name": "standard-profile",
				}
			},
		},
		{
			path: "spec.runtime.customResources",
			mutate: func(runtime map[string]any) {
				runtime["customResources"] = nil
			},
		},
		{
			path: "spec.runtime.customResources.cpuRequestMilliCPU",
			mutate: func(runtime map[string]any) {
				resources := runtime["customResources"].(map[string]any)
				resources["cpuRequestMilliCPU"] = nil
			},
		},
		{
			path: "spec.runtime.customResources.gpuRequest",
			mutate: func(runtime map[string]any) {
				resources := runtime["customResources"].(map[string]any)
				resources["gpuRequest"] = nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			spec := validWorkerTemplateSpec()
			spec.Runtime.ResourceProfileRef = nil
			spec.Runtime.CustomResources = &workerspec.ResourceRequestsLimits{
				CPURequestMilliCPU: 500,
				CPULimitMilliCPU:   1_000,
				MemoryRequestBytes: 1 << 30,
				MemoryLimitBytes:   2 << 30,
			}
			manifest := workerSchemaManifest(t, KindWorkerTemplate, spec)
			var object map[string]any
			require.NoError(t, json.Unmarshal(manifest.Spec, &object))
			test.mutate(object["runtime"].(map[string]any))
			manifest.Spec, _ = json.Marshal(object)

			registry := NewRegistry()
			require.NoError(t, RegisterWorkerSchemas(registry))
			_, err := registry.DecodeAndValidate(manifest)
			require.ErrorIs(t, err, ErrTypedJSONType)
			require.Contains(t, err.Error(), "at path "+test.path)
		})
	}
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
