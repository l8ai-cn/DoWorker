package workerruntime

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type RuntimeImage struct {
	ID      int64
	Digest  string
	Enabled bool
}

type ComputeTarget struct {
	ID      int64
	Kind    workerspec.ComputeTargetKind
	Enabled bool
}

type ResourceProfile struct {
	ID        int64
	Resources workerspec.ResourceRequestsLimits
	Enabled   bool
}

type Request struct {
	OrganizationID    int64
	WorkerTypeSlug    slugkit.Slug
	RuntimeImageID    int64
	PlacementPolicy   workerspec.PlacementPolicy
	ComputeTargetID   int64
	DeploymentMode    workerspec.DeploymentMode
	ResourceProfileID int64
}

type RepositorySelection struct {
	RuntimeImage              *RuntimeImage
	ComputeTarget             *ComputeTarget
	ResourceProfile           *ResourceProfile
	ImageCompatible           bool
	DeploymentCompatible      bool
	ResourceProfileCompatible bool
}

type Resolved struct {
	RuntimeImage workerspec.RuntimeImage
	Placement    workerspec.Placement
}
