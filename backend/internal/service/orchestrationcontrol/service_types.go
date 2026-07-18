package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"time"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
)

type SourceFormat string

const (
	SourceFormatJSON SourceFormat = "json"
	SourceFormatYAML SourceFormat = "yaml"
)

type ResourceSource struct {
	Format  SourceFormat
	Content []byte
}

type ValidateRequest struct {
	Scope  control.Scope
	Source ResourceSource
}

type ValidationResult struct {
	Target            control.ResourceTarget
	Operation         control.PlanOperation
	CanonicalManifest json.RawMessage
	Issues            []control.PlanIssue
}

type PlanRequest = ValidateRequest

type PlanResult struct {
	ValidationResult
	Plan *control.Plan
}

type ResourceCapabilities struct {
	Exists        bool
	CanViewSource bool
	CanReference  bool
	CanPlan       bool
}

type DraftReference struct {
	Path      string
	Reference orchestrationresource.Reference
}

type TargetPlanInput struct {
	Scope              control.Scope
	Operation          control.PlanOperation
	Manifest           orchestrationresource.Manifest
	TypedSpec          any
	Head               *control.ResourceHead
	CurrentRevision    *control.ResourceRevision
	ResolvedReferences []control.ResolvedReference
}

type TargetPlanOutput struct {
	ArtifactKind    string
	ArtifactJSON    json.RawMessage
	OptionsRevision string
	Issues          []control.PlanIssue
}

type TargetPlanner interface {
	TypeMeta() orchestrationresource.TypeMeta
	References(any) ([]DraftReference, error)
	Plan(context.Context, TargetPlanInput) (TargetPlanOutput, error)
}

type Authorizer interface {
	AuthorizeList(context.Context, control.Scope) error
	AuthorizeCreate(context.Context, control.Scope, control.ResourceTarget) error
	AuthorizeUpdate(context.Context, control.Scope, control.ResourceHead) error
	AuthorizeReference(context.Context, control.Scope, control.ResourceHead) error
}

type ReferenceResolver interface {
	Resolve(
		context.Context,
		control.Scope,
		DraftReference,
	) (control.ResolvedReference, error)
}

type WorkerDefinitionPolicyResolver interface {
	EnvironmentBundlePolicy(string) (workerdefinition.EnvironmentBundlePolicy, bool)
}

type ServiceDeps struct {
	Registry          *orchestrationresource.Registry
	Repository        Repository
	Authorizer        Authorizer
	References        ReferenceResolver
	WorkerDefinitions WorkerDefinitionPolicyResolver
	Planners          []TargetPlanner
	RequiredTypes     []orchestrationresource.TypeMeta
	Clock             func() time.Time
	IDGenerator       func() string
	PlanTTL           time.Duration
}
