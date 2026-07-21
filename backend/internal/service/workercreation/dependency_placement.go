package workercreation

import (
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
)

func buildPlacementResolution(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	runtime *runtimeCatalogResolver,
) (workerdependencyartifact.PlacementResolution, error) {
	image, target, profile, ok := runtime.resolvedRuntime()
	if !ok {
		return workerdependencyartifact.PlacementResolution{}, fmt.Errorf(
			"WorkerTemplate artifact runtime was not resolved",
		)
	}
	if refs.ComputeTarget == nil {
		return workerdependencyartifact.PlacementResolution{}, fmt.Errorf(
			"WorkerTemplate artifact compute target reference is missing",
		)
	}
	if target.ID != spec.Placement.ComputeTarget.ID ||
		target.Kind != spec.Placement.ComputeTarget.Kind {
		return workerdependencyartifact.PlacementResolution{}, fmt.Errorf(
			"WorkerTemplate artifact compute target does not match WorkerSpec",
		)
	}
	targetPin, err := referencePin(
		scope,
		*refs.ComputeTarget,
		spec.Placement.ComputeTarget.ID,
	)
	if err != nil {
		return workerdependencyartifact.PlacementResolution{}, err
	}
	var profilePin *workerdependencyartifact.ResourceResolution
	if profile != nil {
		if refs.ResourceProfile == nil {
			return workerdependencyartifact.PlacementResolution{}, fmt.Errorf(
				"WorkerTemplate artifact resource profile reference is missing",
			)
		}
		pin, err := referencePin(scope, *refs.ResourceProfile, profile.ID)
		if err != nil {
			return workerdependencyartifact.PlacementResolution{}, err
		}
		profilePin = &pin
	}
	return workerdependencyartifact.PlacementResolution{
		CatalogRevision: runtime.catalog.Revision(),
		RuntimeImageID:  image.ID,
		ImageReference:  image.Reference,
		ImageDigest:     image.Digest,
		ComputeTarget:   targetPin,
		ResourceProfile: profilePin,
		Spec:            spec.Placement,
	}, nil
}
