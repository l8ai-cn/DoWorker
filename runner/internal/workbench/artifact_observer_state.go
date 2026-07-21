package workbench

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

const (
	artifactObserverStatePath   = ".agent-cloud/workbench/artifact-observer-state.json"
	artifactObserverStateSchema = "agentcloud.artifact-observer-state/v1"
	maxArtifactObserverState    = 4 << 20
)

type artifactObserverState struct {
	SchemaVersion string                                   `json:"schema_version"`
	Files         map[string]artifactObserverFileState     `json:"files"`
	Declarations  map[string]artifactObserverDeclaredState `json:"declarations"`
}

type artifactObserverFileState struct {
	Filename  string `json:"filename"`
	MediaType string `json:"media_type"`
	Digest    string `json:"digest"`
	ByteSize  uint64 `json:"byte_size"`
	Revision  uint64 `json:"revision"`
	Deleted   bool   `json:"deleted,omitempty"`
}

type artifactObserverDeclaredState struct {
	Revision    uint64                      `json:"revision"`
	Fingerprint string                      `json:"fingerprint"`
	Producer    artifactDeclarationProducer `json:"producer"`
	Emitted     bool                        `json:"emitted,omitempty"`
}

func loadArtifactObserverState(root string) (*artifactObserverState, bool, error) {
	workspace, err := os.OpenRoot(root)
	if err != nil {
		return nil, false, err
	}
	defer workspace.Close()
	raw, err := workspace.ReadFile(filepath.FromSlash(artifactObserverStatePath))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if len(raw) > maxArtifactObserverState {
		return nil, false, fmt.Errorf(
			"artifact observer state exceeds %d bytes",
			maxArtifactObserverState,
		)
	}
	var state artifactObserverState
	if err := decodeStrictJSON(raw, &state); err != nil {
		return nil, false, err
	}
	if state.SchemaVersion != artifactObserverStateSchema {
		return nil, false, fmt.Errorf(
			"unsupported artifact observer state schema %q",
			state.SchemaVersion,
		)
	}
	if state.Files == nil || state.Declarations == nil {
		return nil, false, fmt.Errorf("artifact observer state maps are required")
	}
	return &state, true, nil
}

func writeArtifactObserverState(root string, state *artifactObserverState) error {
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if len(raw) > maxArtifactObserverState {
		return fmt.Errorf("artifact observer state exceeds %d bytes", maxArtifactObserverState)
	}
	workspace, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer workspace.Close()
	directory := path.Dir(artifactObserverStatePath)
	if err := workspace.MkdirAll(filepath.FromSlash(directory), 0o755); err != nil {
		return err
	}
	tempPath := path.Join(directory, ".artifact-observer-"+uuid.NewString()+".tmp")
	temp, err := workspace.OpenFile(
		filepath.FromSlash(tempPath),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o644,
	)
	if err != nil {
		return err
	}
	defer workspace.Remove(filepath.FromSlash(tempPath))
	if _, err := temp.Write(raw); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := workspace.Rename(
		filepath.FromSlash(tempPath),
		filepath.FromSlash(artifactObserverStatePath),
	); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	stateDirectory, err := workspace.Open(filepath.FromSlash(directory))
	if err != nil {
		return err
	}
	defer stateDirectory.Close()
	return stateDirectory.Sync()
}
