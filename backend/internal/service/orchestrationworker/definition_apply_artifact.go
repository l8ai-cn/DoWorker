package orchestrationworker

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

func decodeDefinitionApplyArtifact(
	source json.RawMessage,
) (DefinitionApplyArtifact, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var artifact DefinitionApplyArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return DefinitionApplyArtifact{}, control.ErrCorrupt
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return DefinitionApplyArtifact{}, control.ErrCorrupt
	}
	canonical, err := control.CanonicalJSONObject(artifact)
	if err != nil || !bytes.Equal(canonical, source) ||
		artifact.WorkerSpecSnapshotID <= 0 {
		return DefinitionApplyArtifact{}, control.ErrCorrupt
	}
	return artifact, nil
}
