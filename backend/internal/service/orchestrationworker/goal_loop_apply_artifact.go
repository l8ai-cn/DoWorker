package orchestrationworker

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type GoalLoopApplyArtifact struct {
	WorkerSpecSnapshotID int64  `json:"workerSpecSnapshotId"`
	LoopProgramDigest    string `json:"loopProgramDigest,omitempty"`
}

func definitionApplyArtifact(
	kind string,
	snapshotID int64,
	value any,
) ([]byte, error) {
	if kind != resource.KindGoalLoop {
		return control.CanonicalJSONObject(DefinitionApplyArtifact{
			WorkerSpecSnapshotID: snapshotID,
		})
	}
	spec, ok := value.(*resource.GoalLoopResourceSpec)
	if !ok || spec == nil {
		return nil, control.ErrCorrupt
	}
	digest, err := goalLoopProgramDigest(spec.LoopProgram)
	if err != nil {
		return nil, control.ErrCorrupt
	}
	return control.CanonicalJSONObject(GoalLoopApplyArtifact{
		WorkerSpecSnapshotID: snapshotID,
		LoopProgramDigest:    digest,
	})
}

func decodeGoalLoopApplyArtifact(
	source json.RawMessage,
) (GoalLoopApplyArtifact, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var artifact GoalLoopApplyArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return GoalLoopApplyArtifact{}, control.ErrCorrupt
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return GoalLoopApplyArtifact{}, control.ErrCorrupt
	}
	canonical, err := control.CanonicalJSONObject(artifact)
	if err != nil || !bytes.Equal(canonical, source) ||
		artifact.WorkerSpecSnapshotID <= 0 {
		return GoalLoopApplyArtifact{}, control.ErrCorrupt
	}
	return artifact, nil
}

func validateGoalLoopApplyArtifact(
	spec *resource.GoalLoopResourceSpec,
	artifact GoalLoopApplyArtifact,
) error {
	digest, err := goalLoopProgramDigest(spec.LoopProgram)
	if err != nil || artifact.LoopProgramDigest != digest {
		return control.ErrCorrupt
	}
	return nil
}

func goalLoopProgramDigest(
	snapshot *resource.GoalLoopProgramSnapshot,
) (string, error) {
	if snapshot == nil {
		return "", nil
	}
	canonical, err := control.CanonicalJSONObject(snapshot)
	if err != nil {
		return "", err
	}
	return control.DigestCanonicalJSON(canonical)
}
