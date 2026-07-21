package workerdependencyartifact

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

const PlanArtifactKind = "WorkerTemplateBuild"

const planArtifactVersion uint16 = 1

type planArtifactDocument struct {
	Version                    uint16          `json:"version"`
	WorkerSpec                 json.RawMessage `json:"workerSpec"`
	ResolvedDependencies       json.RawMessage `json:"resolvedDependencies"`
	ResolvedDependenciesDigest string          `json:"resolvedDependenciesDigest"`
}

type DecodedPlanArtifact struct {
	WorkerSpec                 workerspec.Spec
	WorkerSpecJSON             []byte
	ResolvedDependencies       workerdependency.Document
	ResolvedDependenciesJSON   []byte
	ResolvedDependenciesDigest string
}

func encodePlanArtifact(
	spec workerspec.Spec,
	dependenciesJSON []byte,
	dependenciesDigest string,
) ([]byte, string, error) {
	specJSON, err := encodePlanWorkerSpec(spec)
	if err != nil {
		return nil, "", err
	}
	encoded, err := control.CanonicalJSONObject(planArtifactDocument{
		Version:                    planArtifactVersion,
		WorkerSpec:                 json.RawMessage(specJSON),
		ResolvedDependencies:       json.RawMessage(dependenciesJSON),
		ResolvedDependenciesDigest: dependenciesDigest,
	})
	if err != nil {
		return nil, "", fmt.Errorf("encode WorkerTemplate build artifact: %w", err)
	}
	digest, err := control.DigestCanonicalJSON(encoded)
	if err != nil {
		return nil, "", fmt.Errorf("digest WorkerTemplate build artifact: %w", err)
	}
	return encoded, digest, nil
}

func DecodePlanArtifact(data []byte) (DecodedPlanArtifact, error) {
	canonical, err := control.CanonicalJSONObject(data)
	if err != nil || !bytes.Equal(canonical, data) {
		return DecodedPlanArtifact{}, fmt.Errorf(
			"WorkerTemplate build artifact must be canonical JSON",
		)
	}
	var document planArtifactDocument
	if err := decodePlanArtifactStrict(canonical, &document); err != nil {
		return DecodedPlanArtifact{}, fmt.Errorf(
			"decode WorkerTemplate build artifact: %w",
			err,
		)
	}
	if document.Version != planArtifactVersion {
		return DecodedPlanArtifact{}, fmt.Errorf(
			"WorkerTemplate build artifact version %d is unsupported",
			document.Version,
		)
	}
	if len(document.WorkerSpec) == 0 || len(document.ResolvedDependencies) == 0 {
		return DecodedPlanArtifact{}, fmt.Errorf(
			"WorkerTemplate build artifact is incomplete",
		)
	}
	spec, specJSON, specDigest, err := decodePlanWorkerSpec(document.WorkerSpec)
	if err != nil {
		return DecodedPlanArtifact{}, err
	}
	dependencies, err := workerdependency.Decode(document.ResolvedDependencies)
	if err != nil {
		return DecodedPlanArtifact{}, fmt.Errorf(
			"decode resolved Worker dependencies: %w",
			err,
		)
	}
	dependenciesJSON, digest, err := workerdependency.EncodeAndDigest(dependencies)
	if err != nil {
		return DecodedPlanArtifact{}, err
	}
	if digest != document.ResolvedDependenciesDigest ||
		dependencies.Worker.SpecDigest != specDigest {
		return DecodedPlanArtifact{}, fmt.Errorf(
			"WorkerTemplate build artifact digest binding is invalid",
		)
	}
	if err := ValidateWorkerSpecConsistency(spec, dependencies); err != nil {
		return DecodedPlanArtifact{}, err
	}
	return DecodedPlanArtifact{
		WorkerSpec: spec, WorkerSpecJSON: bytes.Clone(specJSON),
		ResolvedDependencies:       dependencies,
		ResolvedDependenciesJSON:   bytes.Clone(dependenciesJSON),
		ResolvedDependenciesDigest: digest,
	}, nil
}

func decodePlanArtifactStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	err := decoder.Decode(&trailing)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("trailing JSON data")
	}
	return err
}
