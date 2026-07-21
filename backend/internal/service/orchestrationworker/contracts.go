package orchestrationworker

import (
	"context"
	"encoding/json"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
)

type BindingResolver interface {
	ResolveEntityID(
		context.Context,
		control.Scope,
		control.ResolvedReference,
	) (int64, error)
	ResolveToolModel(
		context.Context,
		control.Scope,
		control.ResolvedReference,
	) (ToolModelBindingResolution, error)
}

type ToolModelBindingResolution struct {
	Binding         control.ResolvedReference
	ModelBinding    control.ResolvedReference
	ModelResourceID int64
}

type DefinitionResolver interface {
	ResolveWorkerSpecSnapshotID(
		context.Context,
		control.Scope,
		control.ResolvedReference,
	) (int64, error)
	ResolvePromptSpec(
		context.Context,
		control.Scope,
		control.ResolvedReference,
	) (resource.PromptSpec, error)
}

type GoalLoopSlugChecker interface {
	ExistsSlug(context.Context, int64, string) (bool, error)
}

type WorkerCompilation struct {
	ArtifactJSON json.RawMessage
	Issues       []control.PlanIssue
}

type WorkerCompiler interface {
	Revision() string
	Compile(
		context.Context,
		control.Scope,
		workercreation.Draft,
	) (WorkerCompilation, error)
}
