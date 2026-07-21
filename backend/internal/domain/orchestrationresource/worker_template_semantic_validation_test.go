package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func TestWorkerTemplateRejectsInvalidStaticWorkerSpecSemantics(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		mutate func(*WorkerTemplateSpec)
	}{
		{
			name:  "runtime image",
			field: "runtime image id",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Runtime.RuntimeImageID = 0
			},
		},
		{
			name:  "placement policy",
			field: "placement policy",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Runtime.PlacementPolicy = "random"
			},
		},
		{
			name:  "deployment mode",
			field: "deployment mode",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Runtime.DeploymentMode = "shared"
			},
		},
		{
			name:  "missing resource profile",
			field: "resource profile id",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Runtime.ResourceProfileRef = nil
			},
		},
		{
			name:  "type config schema",
			field: "schema version",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.SchemaVersion = 0
			},
		},
		{
			name:  "nil values",
			field: "null is not allowed at path spec.typeConfig.values",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.Values = nil
			},
		},
		{
			name:  "nil secret refs",
			field: "null is not allowed at path spec.typeConfig.secretRefs",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.SecretRefs = nil
			},
		},
		{
			name:  "value and secret conflict",
			field: "cannot appear in both values and secret refs",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.Values["api-token"] = "plaintext"
			},
		},
		{
			name:  "interaction mode",
			field: "interaction mode",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.InteractionMode = "chat"
			},
		},
		{
			name:  "automation level",
			field: "automation level",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.TypeConfig.AutomationLevel = "supervised"
			},
		},
		{
			name:  "branch without repository",
			field: "repository is required",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.RepositoryRef = nil
			},
		},
		{
			name:  "repository without branch",
			field: "branch is required",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.Branch = ""
			},
		},
		{
			name:  "unnormalized branch",
			field: "branch must be normalized",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.Branch = " main"
			},
		},
		{
			name:  "long branch",
			field: "branch exceeds",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.Branch = strings.Repeat("a", 256)
			},
		},
		{
			name:  "knowledge mode",
			field: "invalid mode",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Workspace.KnowledgeMounts[0].Mode = "write-only"
			},
		},
		{
			name:  "idle timeout",
			field: "idle timeout must be positive",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Lifecycle.IdleTimeoutMinutes = 0
			},
		},
		{
			name:  "manual timeout",
			field: "idle timeout must be zero",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Lifecycle.TerminationPolicy = workerspec.TerminationPolicyManual
			},
		},
		{
			name:  "termination policy",
			field: "termination policy",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Lifecycle.TerminationPolicy = "never"
			},
		},
		{
			name:  "alias",
			field: "metadata alias exceeds",
			mutate: func(spec *WorkerTemplateSpec) {
				spec.Metadata.Alias = strings.Repeat("a", 101)
			},
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

func TestWorkerTemplateRejectsInvalidCustomResources(t *testing.T) {
	spec := validWorkerTemplateSpec()
	spec.Runtime.ResourceProfileRef = nil
	spec.Runtime.CustomResources = &workerspec.ResourceRequestsLimits{
		CPURequestMilliCPU:  2000,
		CPULimitMilliCPU:    1000,
		MemoryRequestBytes:  1024,
		MemoryLimitBytes:    2048,
		StorageRequestBytes: 1024,
		StorageLimitBytes:   2048,
	}

	requireWorkerTemplateError(t, spec, "cpu request must not exceed cpu limit")
}
