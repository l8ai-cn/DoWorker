package orchestrationworker

import (
	"context"
	"encoding/json"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
)

type BindingResolver interface {
	ResolveEntityID(
		context.Context,
		control.Scope,
		control.ResolvedReference,
	) (int64, error)
	ResolveToolModelResourceID(
		context.Context,
		control.Scope,
		control.ResolvedReference,
	) (int64, error)
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
