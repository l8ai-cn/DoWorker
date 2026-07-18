package workerdependencyartifact

import (
	"bytes"
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
)

type Input struct {
	Scope          control.Scope
	Definition     workerdefinition.Definition
	AgentfileLayer string
	PlanReferences []control.ResolvedReference
	WorkerSpec     workerspec.Spec
	Dependencies   ResolvedDependencies
}

type Artifact struct {
	documentJSON []byte
	digest       string
	workerSpec   []byte
	specDigest   string
	planJSON     []byte
	planDigest   string
}

func Build(input Input) (Artifact, error) {
	if err := validateScope(input); err != nil {
		return Artifact{}, err
	}
	if err := validateBuildInputBudget(input); err != nil {
		return Artifact{}, err
	}
	spec, specJSON, specDigest, err := canonicalWorkerSpec(input.WorkerSpec)
	if err != nil {
		return Artifact{}, err
	}
	documentInput, err := materializeDocument(
		input.Scope,
		input.Definition,
		input.AgentfileLayer,
		spec,
		specDigest,
		input.Dependencies,
	)
	if err != nil {
		return Artifact{}, err
	}
	document, err := workerdependency.NormalizeAndValidate(documentInput)
	if err != nil {
		return Artifact{}, fmt.Errorf("validate worker dependency document: %w", err)
	}
	if err := validateWorkerSpecConsistency(spec, document); err != nil {
		return Artifact{}, err
	}
	if err := validateDefinitionDependencies(
		input.Scope,
		input.Definition,
		spec,
		document,
	); err != nil {
		return Artifact{}, err
	}
	if err := validateReferenceClosure(
		input.Scope,
		input.PlanReferences,
		input.Dependencies.ToolModels,
		document,
	); err != nil {
		return Artifact{}, err
	}
	encoded, digest, err := workerdependency.EncodeAndDigest(document)
	if err != nil {
		return Artifact{}, fmt.Errorf("encode worker dependency artifact: %w", err)
	}
	planJSON, planDigest, err := encodePlanArtifact(spec, encoded, digest)
	if err != nil {
		return Artifact{}, err
	}
	return Artifact{
		documentJSON: bytes.Clone(encoded),
		digest:       digest,
		workerSpec:   bytes.Clone(specJSON),
		specDigest:   specDigest,
		planJSON:     bytes.Clone(planJSON),
		planDigest:   planDigest,
	}, nil
}

func (artifact Artifact) JSON() []byte {
	return bytes.Clone(artifact.documentJSON)
}

func (artifact Artifact) Digest() string {
	return artifact.digest
}

func (artifact Artifact) WorkerSpecJSON() []byte {
	return bytes.Clone(artifact.workerSpec)
}

func (artifact Artifact) WorkerSpecDigest() string {
	return artifact.specDigest
}

func (artifact Artifact) PlanJSON() []byte {
	return bytes.Clone(artifact.planJSON)
}

func (artifact Artifact) PlanDigest() string {
	return artifact.planDigest
}

func validateScope(input Input) error {
	return input.Scope.Validate()
}

func canonicalWorkerSpec(
	input workerspec.Spec,
) (workerspec.Spec, []byte, string, error) {
	spec, err := workerspec.NormalizeAndValidate(input)
	if err != nil {
		return workerspec.Spec{}, nil, "", fmt.Errorf(
			"validate Plan WorkerSpec: %w",
			err,
		)
	}
	encoded, err := workerspec.EncodeSpec(spec)
	if err != nil {
		return workerspec.Spec{}, nil, "", fmt.Errorf("encode Plan WorkerSpec: %w", err)
	}
	canonical, err := control.CanonicalJSONObject(encoded)
	if err != nil {
		return workerspec.Spec{}, nil, "", fmt.Errorf(
			"canonicalize Plan WorkerSpec: %w",
			err,
		)
	}
	digest, err := control.DigestCanonicalJSON(canonical)
	if err != nil {
		return workerspec.Spec{}, nil, "", fmt.Errorf("digest Plan WorkerSpec: %w", err)
	}
	return spec, canonical, digest, nil
}
