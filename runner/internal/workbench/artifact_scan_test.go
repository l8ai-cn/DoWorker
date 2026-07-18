package workbench

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArtifactObserverDiscovers3DModelDeliverables(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)

	expected := map[string]string{
		"model.3mf":   "model/3mf",
		"model.blend": "application/x-blender",
		"model.glb":   "model/gltf-binary",
		"model.gltf":  "model/gltf+json",
		"model.scad":  "text/plain",
		"model.step":  "model/step",
		"model.stl":   "model/stl",
	}
	for filename := range expected {
		writeArtifactFile(t, root, "deliverables/"+filename, "model")
	}

	artifacts, err := observer.Scan()

	require.NoError(t, err)
	require.Len(t, artifacts, len(expected))
	for _, artifact := range artifacts {
		require.Equal(t, expected[artifact.GetFilename()], artifact.GetMediaType())
	}
}
