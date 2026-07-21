package workerdependencyartifact

import (
	"bytes"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type planWorkerSpec struct {
	Version    workerspec.Version    `json:"version"`
	Runtime    planWorkerRuntime     `json:"runtime"`
	Placement  workerspec.Placement  `json:"placement"`
	TypeConfig workerspec.TypeConfig `json:"type_config"`
	Workspace  workerspec.Workspace  `json:"workspace"`
	Lifecycle  workerspec.Lifecycle  `json:"lifecycle"`
	Metadata   workerspec.Metadata   `json:"metadata"`
}

type planWorkerRuntime struct {
	ModelBinding      workerspec.ModelBinding `json:"model_binding"`
	ToolModelBindings []planToolModelBinding  `json:"tool_model_bindings,omitempty"`
	WorkerType        workerspec.WorkerType   `json:"worker_type"`
	Image             workerspec.RuntimeImage `json:"image"`
}

type planToolModelBinding struct {
	Role         slugkit.Slug             `json:"role"`
	ModelBinding workerspec.ModelBinding  `json:"model_binding"`
	Modality     airesource.Modality      `json:"modality"`
	Capability   airesource.Capability    `json:"capability"`
	Environment  planToolModelEnvironment `json:"environment"`
}

type planToolModelEnvironment struct {
	APIKeyTarget  string `json:"api_key_target"`
	BaseURLTarget string `json:"base_url_target"`
	ModelIDTarget string `json:"model_id_target"`
}

func encodePlanWorkerSpec(spec workerspec.Spec) ([]byte, error) {
	document := planWorkerSpecFromDomain(spec)
	encoded, err := control.CanonicalJSONObject(document)
	if err != nil {
		return nil, fmt.Errorf("encode planned WorkerSpec: %w", err)
	}
	return encoded, nil
}

func decodePlanWorkerSpec(
	data []byte,
) (workerspec.Spec, []byte, string, error) {
	var document planWorkerSpec
	if err := decodePlanArtifactStrict(data, &document); err != nil {
		return workerspec.Spec{}, nil, "", fmt.Errorf(
			"decode planned WorkerSpec: %w",
			err,
		)
	}
	spec, encoded, digest, err := canonicalWorkerSpec(document.domain())
	if err != nil {
		return workerspec.Spec{}, nil, "", err
	}
	planEncoded, err := encodePlanWorkerSpec(spec)
	if err != nil {
		return workerspec.Spec{}, nil, "", err
	}
	if !bytes.Equal(planEncoded, data) {
		return workerspec.Spec{}, nil, "", fmt.Errorf(
			"planned WorkerSpec must be canonical JSON",
		)
	}
	return spec, encoded, digest, nil
}

func planWorkerSpecFromDomain(spec workerspec.Spec) planWorkerSpec {
	tools := make([]planToolModelBinding, len(spec.Runtime.ToolModelBindings))
	for index, binding := range spec.Runtime.ToolModelBindings {
		tools[index] = planToolModelBinding{
			Role: binding.Role, ModelBinding: binding.ModelBinding,
			Modality: binding.Modality, Capability: binding.Capability,
			Environment: planToolModelEnvironment{
				APIKeyTarget:  binding.Environment.APIKey,
				BaseURLTarget: binding.Environment.BaseURL,
				ModelIDTarget: binding.Environment.ModelID,
			},
		}
	}
	return planWorkerSpec{
		Version: spec.Version,
		Runtime: planWorkerRuntime{
			ModelBinding: spec.Runtime.ModelBinding, ToolModelBindings: tools,
			WorkerType: spec.Runtime.WorkerType, Image: spec.Runtime.Image,
		},
		Placement: spec.Placement, TypeConfig: spec.TypeConfig,
		Workspace: spec.Workspace, Lifecycle: spec.Lifecycle, Metadata: spec.Metadata,
	}
}

func (document planWorkerSpec) domain() workerspec.Spec {
	tools := make([]workerspec.ToolModelBinding, len(document.Runtime.ToolModelBindings))
	for index, binding := range document.Runtime.ToolModelBindings {
		tools[index] = workerspec.ToolModelBinding{
			Role: binding.Role, ModelBinding: binding.ModelBinding,
			Modality: binding.Modality, Capability: binding.Capability,
			Environment: workerspec.ToolModelEnvironment{
				APIKey:  binding.Environment.APIKeyTarget,
				BaseURL: binding.Environment.BaseURLTarget,
				ModelID: binding.Environment.ModelIDTarget,
			},
		}
	}
	return workerspec.Spec{
		Version: document.Version,
		Runtime: workerspec.Runtime{
			ModelBinding: document.Runtime.ModelBinding, ToolModelBindings: tools,
			WorkerType: document.Runtime.WorkerType, Image: document.Runtime.Image,
		},
		Placement: document.Placement, TypeConfig: document.TypeConfig,
		Workspace: document.Workspace, Lifecycle: document.Lifecycle,
		Metadata: document.Metadata,
	}
}
