package workerdependencyartifact

import (
	"bytes"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
)

type ApplyArtifact struct {
	workerSpecJSON     []byte
	dependenciesJSON   []byte
	dependenciesDigest string
	secretReferences   []workerdependency.SecretReference
}

func DecodeApplyPlan(plan control.Plan) (ApplyArtifact, error) {
	if err := plan.Validate(); err != nil {
		return ApplyArtifact{}, err
	}
	if plan.Target.Kind != resource.KindWorkerTemplate ||
		plan.ArtifactKind != PlanArtifactKind {
		return ApplyArtifact{}, fmt.Errorf("Plan does not contain a WorkerTemplate build")
	}
	digest, err := control.DigestCanonicalJSON(plan.ArtifactJSON)
	if err != nil || digest != plan.ArtifactDigest {
		return ApplyArtifact{}, fmt.Errorf("Plan artifact digest is invalid")
	}
	decoded, err := DecodePlanArtifact(plan.ArtifactJSON)
	if err != nil {
		return ApplyArtifact{}, err
	}
	if decoded.ResolvedDependencies.OrganizationID != plan.Scope.OrganizationID ||
		decoded.ResolvedDependencies.Namespace != plan.Scope.OrganizationSlug {
		return ApplyArtifact{}, fmt.Errorf("Plan scope does not match Worker dependencies")
	}
	if err := validateDirectReferenceClosure(
		plan.Scope,
		plan.ResolvedReferences,
		decoded.ResolvedDependencies,
	); err != nil {
		return ApplyArtifact{}, err
	}
	return ApplyArtifact{
		workerSpecJSON:     bytes.Clone(decoded.WorkerSpecJSON),
		dependenciesJSON:   bytes.Clone(decoded.ResolvedDependenciesJSON),
		dependenciesDigest: decoded.ResolvedDependenciesDigest,
		secretReferences: cloneSecretReferences(
			decoded.ResolvedDependencies.SecretReferences,
		),
	}, nil
}

func (artifact ApplyArtifact) WorkerSpecJSON() []byte {
	return bytes.Clone(artifact.workerSpecJSON)
}

func (artifact ApplyArtifact) DependenciesJSON() []byte {
	return bytes.Clone(artifact.dependenciesJSON)
}

func (artifact ApplyArtifact) DependenciesDigest() string {
	return artifact.dependenciesDigest
}

func (artifact ApplyArtifact) SecretReferences() []workerdependency.SecretReference {
	return cloneSecretReferences(artifact.secretReferences)
}

func cloneSecretReferences(
	references []workerdependency.SecretReference,
) []workerdependency.SecretReference {
	result := append([]workerdependency.SecretReference{}, references...)
	return result
}
