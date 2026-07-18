package workerspec

import (
	"context"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type Scope struct {
	OrgID  int64
	UserID int64
}

type RuntimeSelection struct {
	RuntimeImageID    int64
	PlacementPolicy   domain.PlacementPolicy
	ComputeTargetID   int64
	DeploymentMode    domain.DeploymentMode
	ResourceProfileID int64
	CustomResources   *domain.ResourceRequestsLimits
}

type Draft struct {
	ModelResourceID      int64
	ToolModelResourceIDs map[string]int64
	WorkerTypeSlug       slugkit.Slug
	Runtime              RuntimeSelection
	TypeConfig           domain.TypeConfig
	Workspace            domain.Workspace
	Lifecycle            domain.Lifecycle
	Metadata             domain.Metadata
}

type WorkerTypeResolution struct {
	WorkerType                domain.WorkerType
	TypeSchema                domain.TypeSchema
	SupportedInteractionModes []domain.InteractionMode
	ModelRequirement          domain.ModelRequirement
	ToolModelRequirements     []domain.ToolModelRequirement
}

type WorkerTypeResolver interface {
	ResolveWorkerType(
		ctx context.Context,
		scope Scope,
		slug slugkit.Slug,
	) (WorkerTypeResolution, error)
}

type RuntimeResolver interface {
	ResolveRuntime(
		ctx context.Context,
		scope Scope,
		workerType slugkit.Slug,
		selection RuntimeSelection,
	) (workerruntime.Resolved, error)
}

type ModelResolver interface {
	ResolveModel(
		ctx context.Context,
		scope Scope,
		requirement domain.ModelRequirement,
		resourceID int64,
	) (domain.ModelBinding, error)
}

type ToolModelResolver interface {
	ResolveToolModel(
		ctx context.Context,
		scope Scope,
		requirement domain.ToolModelRequirement,
		resourceID int64,
	) (domain.ToolModelBinding, error)
}

type SecretReferenceResolver interface {
	ResolveSecretReference(
		ctx context.Context,
		scope Scope,
		workerType slugkit.Slug,
		field string,
		reference domain.SecretReference,
	) error
}

type WorkspaceResolver interface {
	ResolveWorkspace(
		ctx context.Context,
		scope Scope,
		workerType slugkit.Slug,
		workspace domain.Workspace,
	) (domain.Workspace, error)
}

type SnapshotRepository interface {
	Create(
		ctx context.Context,
		resolved ResolvedSnapshot,
	) (domain.Snapshot, error)
	GetByID(
		ctx context.Context,
		organizationID, snapshotID int64,
	) (domain.Snapshot, error)
	GetByIDs(
		ctx context.Context,
		organizationID int64,
		snapshotIDs []int64,
	) ([]domain.Snapshot, error)
	ListByOrganization(
		ctx context.Context,
		organizationID int64,
	) ([]domain.Snapshot, error)
	Delete(
		ctx context.Context,
		organizationID, snapshotID int64,
	) error
}

type ResolverDeps struct {
	WorkerTypes WorkerTypeResolver
	Runtime     RuntimeResolver
	Models      ModelResolver
	ToolModels  ToolModelResolver
	Secrets     SecretReferenceResolver
	Workspaces  WorkspaceResolver
}
