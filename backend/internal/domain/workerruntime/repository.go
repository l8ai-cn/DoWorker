package workerruntime

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type Repository interface {
	GetRuntimeImageByIDForOrganization(
		ctx context.Context,
		organizationID, imageID int64,
	) (*RuntimeImage, error)
	GetComputeTargetByIDForOrganization(
		ctx context.Context,
		organizationID, targetID int64,
	) (*ComputeTarget, error)
	GetResourceProfileByIDForOrganization(
		ctx context.Context,
		organizationID, profileID int64,
	) (*ResourceProfile, error)
	IsRuntimeImageCompatibleWithWorkerType(
		ctx context.Context,
		organizationID int64,
		workerType slugkit.Slug,
		imageID int64,
	) (bool, error)
	IsComputeTargetCompatibleWithDeployment(
		ctx context.Context,
		organizationID, targetID int64,
		mode workerspec.DeploymentMode,
	) (bool, error)
	IsComputeTargetCompatibleWithResourceProfile(
		ctx context.Context,
		organizationID, targetID, profileID int64,
	) (bool, error)
}
