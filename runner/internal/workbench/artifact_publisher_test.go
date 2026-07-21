package workbench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishArtifactDeclarationWritesValidatedArtifactAtomically(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "output"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "output", "demo.mp4"),
		[]byte(validMP4Fixture("video")),
		0o644,
	))

	published, err := PublishArtifactDeclaration(
		root,
		json.RawMessage(validPublishedVideoDeclaration(1)),
	)

	require.NoError(t, err)
	require.Equal(t, "demo-video", published.ArtifactID)
	require.Equal(t, uint64(1), published.Revision)
	require.Equal(
		t,
		".agent-cloud/workbench/artifacts/demo-video.json",
		published.DeclarationPath,
	)
	observer, observerErr := NewArtifactObserver(root)
	require.NoError(t, observerErr)
	require.Empty(t, mustScanArtifacts(t, observer))
}

func TestPublishArtifactDeclarationRejectsInvalidRevisionWithoutReplacingCurrent(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "output"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "output", "demo.mp4"),
		[]byte(validMP4Fixture("video")),
		0o644,
	))
	_, err := PublishArtifactDeclaration(
		root,
		json.RawMessage(validPublishedVideoDeclaration(1)),
	)
	require.NoError(t, err)
	path := filepath.Join(
		root,
		filepath.FromSlash(".agent-cloud/workbench/artifacts/demo-video.json"),
	)
	before, err := os.ReadFile(path)
	require.NoError(t, err)

	_, err = PublishArtifactDeclaration(
		root,
		json.RawMessage(validPublishedVideoDeclaration(3)),
	)

	require.ErrorContains(t, err, "changed artifact revision must be 2")
	after, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	require.Equal(t, before, after)
}

func TestPublishArtifactDeclarationAllowsChangedFileAtNextRevision(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "output", "demo.mp4")
	require.NoError(t, os.MkdirAll(filepath.Dir(output), 0o755))
	require.NoError(t, os.WriteFile(
		output,
		[]byte(validMP4Fixture("version-one")),
		0o644,
	))
	_, err := PublishArtifactDeclaration(
		root,
		json.RawMessage(validPublishedVideoDeclaration(1)),
	)
	require.NoError(t, err)
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		output,
		[]byte(validMP4Fixture("version-two")),
		0o644,
	))

	published, err := PublishArtifactDeclaration(
		root,
		json.RawMessage(validPublishedVideoDeclaration(2)),
	)

	require.NoError(t, err)
	require.Equal(t, uint64(2), published.Revision)
	artifacts := mustScanArtifacts(t, observer)
	require.Len(t, artifacts, 1)
	require.Equal(t, uint64(2), artifacts[0].GetRevision())
}

func TestPublishArtifactDeclarationRejectsDeclarationDirectorySymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Symlink(
		outside,
		filepath.Join(root, ".agent-cloud"),
	))

	_, err := PublishArtifactDeclaration(
		root,
		json.RawMessage(validPublishedVideoDeclaration(1)),
	)

	require.Error(t, err)
	require.Empty(t, mustReadDir(t, outside))
}

func TestPublishArtifactDeclarationRejectsAgentSuppliedToolExecutionID(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "output"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "output", "demo.mp4"),
		[]byte(validMP4Fixture("video")),
		0o644,
	))
	declaration := validPublishedVideoDeclaration(1)
	declaration = strings.Replace(
		declaration,
		`"type":"artifact.publish"`,
		`"type":"artifact.publish","tool_execution_id":"agent-chosen"`,
		1,
	)

	_, err := PublishArtifactDeclaration(root, json.RawMessage(declaration))

	require.ErrorContains(t, err, "producer.tool_execution_id")
}

func TestPublishArtifactDeclarationRejectsAgentSuppliedCommandID(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "output"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "output", "demo.mp4"),
		[]byte(validMP4Fixture("video")),
		0o644,
	))
	declaration := validPublishedVideoDeclaration(1)
	declaration = strings.Replace(
		declaration,
		`"type":"artifact.publish"`,
		`"type":"artifact.publish","command_id":"agent-chosen"`,
		1,
	)

	_, err := PublishArtifactDeclaration(root, json.RawMessage(declaration))

	require.ErrorContains(t, err, "producer.command_id")
}

func TestPublishArtifactDeclarationRejectsNonFinalVideoStage(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "output"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "output", "demo.mp4"),
		[]byte(validMP4Fixture("video")),
		0o644,
	))
	declaration := strings.Replace(
		validPublishedVideoDeclaration(1),
		`"stage":"ready"`,
		`"stage":"rendering"`,
		1,
	)

	_, err := PublishArtifactDeclaration(root, json.RawMessage(declaration))

	require.ErrorContains(t, err, "published video manifest stage must be ready")
}

func validPublishedVideoDeclaration(revision uint64) string {
	return `{
		"schema_version":"agentcloud.agent-workbench.artifact/v1",
		"artifact_id":"demo-video",
		"revision":` + jsonNumber(revision) + `,
		"role":"preview",
		"primary_representation_id":"playable",
		"producer":{"namespace":"agentcloud.mcp","type":"artifact.publish"},
		"representations":[{
			"representation_id":"playable",
			"path":"output/demo.mp4",
			"media_type":"video/mp4",
			"role":"playable"
		}],
		"manifest":{
			"kind":"video",
			"stage":"ready",
			"playable_representation_id":"playable"
		}
	}`
}

func jsonNumber(value uint64) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}

func mustScanArtifacts(t *testing.T, observer *ArtifactObserver) []*ArtifactDescriptor {
	t.Helper()
	artifacts, err := observer.Scan()
	require.NoError(t, err)
	return artifacts
}

func mustReadDir(t *testing.T, path string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(path)
	require.NoError(t, err)
	return entries
}
