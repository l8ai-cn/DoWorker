package orchestrationresource

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type WorkerTemplateSpec struct {
	OptionsRevision string                       `json:"optionsRevision" yaml:"optionsRevision"`
	WorkerType      slugkit.Slug                 `json:"workerType" yaml:"workerType"`
	ModelRef        *Reference                   `json:"modelRef,omitempty" yaml:"modelRef,omitempty"`
	ToolRefs        map[string]Reference         `json:"toolRefs" yaml:"toolRefs"`
	Runtime         WorkerTemplateRuntimeSpec    `json:"runtime" yaml:"runtime"`
	TypeConfig      WorkerTemplateTypeConfigSpec `json:"typeConfig" yaml:"typeConfig"`
	Workspace       WorkerTemplateWorkspaceSpec  `json:"workspace" yaml:"workspace"`
	Lifecycle       WorkerTemplateLifecycleSpec  `json:"lifecycle" yaml:"lifecycle"`
	Metadata        WorkerTemplateMetadataSpec   `json:"metadata" yaml:"metadata"`
}

type WorkerTemplateRuntimeSpec struct {
	RuntimeImageID     int64                              `json:"runtimeImageId" yaml:"runtimeImageId"`
	PlacementPolicy    workerspec.PlacementPolicy         `json:"placementPolicy" yaml:"placementPolicy"`
	ComputeTargetRef   Reference                          `json:"computeTargetRef" yaml:"computeTargetRef"`
	DeploymentMode     workerspec.DeploymentMode          `json:"deploymentMode" yaml:"deploymentMode"`
	ResourceProfileRef *Reference                         `json:"resourceProfileRef,omitempty" yaml:"resourceProfileRef,omitempty"`
	CustomResources    *workerspec.ResourceRequestsLimits `json:"customResources,omitempty" yaml:"customResources,omitempty"`
}

type WorkerTemplateTypeConfigSpec struct {
	SchemaVersion   uint32                     `json:"schemaVersion" yaml:"schemaVersion"`
	Values          map[string]any             `json:"values" yaml:"values"`
	SecretRefs      map[string]Reference       `json:"secretRefs" yaml:"secretRefs"`
	InteractionMode workerspec.InteractionMode `json:"interactionMode" yaml:"interactionMode"`
	AutomationLevel workerspec.AutomationLevel `json:"automationLevel" yaml:"automationLevel"`
}

type WorkerTemplateWorkspaceSpec struct {
	RepositoryRef          *Reference                            `json:"repositoryRef,omitempty" yaml:"repositoryRef,omitempty"`
	Branch                 string                                `json:"branch" yaml:"branch"`
	SkillRefs              []Reference                           `json:"skillRefs" yaml:"skillRefs"`
	KnowledgeMounts        []WorkerTemplateKnowledgeMount        `json:"knowledgeMounts" yaml:"knowledgeMounts"`
	EnvironmentBundleRefs  []Reference                           `json:"environmentBundleRefs" yaml:"environmentBundleRefs"`
	ConfigDocumentBindings []WorkerTemplateConfigDocumentBinding `json:"configDocumentBindings" yaml:"configDocumentBindings"`
	Instructions           string                                `json:"instructions" yaml:"instructions"`
}

type WorkerTemplateKnowledgeMount struct {
	Ref  Reference                     `json:"ref" yaml:"ref"`
	Mode workerspec.KnowledgeMountMode `json:"mode" yaml:"mode"`
}

type WorkerTemplateConfigDocumentBinding struct {
	DocumentID      string    `json:"documentId" yaml:"documentId"`
	ConfigBundleRef Reference `json:"configBundleRef" yaml:"configBundleRef"`
}

type WorkerTemplateLifecycleSpec struct {
	TerminationPolicy  workerspec.TerminationPolicy `json:"terminationPolicy" yaml:"terminationPolicy"`
	IdleTimeoutMinutes uint32                       `json:"idleTimeoutMinutes" yaml:"idleTimeoutMinutes"`
}

type WorkerTemplateMetadataSpec struct {
	Alias string `json:"alias" yaml:"alias"`
}
