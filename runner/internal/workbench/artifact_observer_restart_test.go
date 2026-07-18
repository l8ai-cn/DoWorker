package workbench

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArtifactObserverRestartEmitsArtifactCreatedAfterBaseline(t *testing.T) {
	root := t.TempDir()
	writeArtifactFile(t, root, "output/historical.png", "historical")
	_, err := NewArtifactObserver(root)
	require.NoError(t, err)

	writeArtifactFile(t, root, "output/restart-result.png", "new")
	restarted, err := NewArtifactObserver(root)
	require.NoError(t, err)

	artifacts, err := restarted.Scan()

	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	require.Equal(
		t,
		"workspace:output/restart-result.png",
		artifacts[0].GetArtifactId(),
	)
}

func TestArtifactObserverRestartDoesNotReplayCheckpointedHistory(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/result.png", "result")
	require.Len(t, mustScanArtifacts(t, observer), 1)
	require.Empty(t, mustScanArtifacts(t, observer))

	restarted, err := NewArtifactObserver(root)
	require.NoError(t, err)

	require.Empty(t, mustScanArtifacts(t, restarted))
}

func TestArtifactObserverRestartReplaysLastUncheckpointedScan(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/result.png", "result")
	first := mustScanArtifacts(t, observer)
	require.Len(t, first, 1)

	restarted, err := NewArtifactObserver(root)
	require.NoError(t, err)
	replayed := mustScanArtifacts(t, restarted)

	require.Len(t, replayed, 1)
	require.Equal(t, first[0].GetArtifactId(), replayed[0].GetArtifactId())
	require.Equal(t, first[0].GetRevision(), replayed[0].GetRevision())
}

func TestArtifactObserverRestartPreservesArtifactRevision(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/result.png", "one")
	first := mustScanArtifacts(t, observer)
	require.Len(t, first, 1)
	require.Empty(t, mustScanArtifacts(t, observer))

	restarted, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/result.png", "two")
	second := mustScanArtifacts(t, restarted)

	require.Len(t, second, 1)
	require.Equal(t, uint64(2), second[0].GetRevision())
}

func TestArtifactObserverRestartRecoversDeclaredArtifact(t *testing.T) {
	root := t.TempDir()
	_, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/result.html", "<h1>result</h1>")
	writeArtifactDeclaration(t, root, "page.json", declaredPageArtifact(1))

	restarted, err := NewArtifactObserver(root)
	require.NoError(t, err)
	artifacts := mustScanArtifacts(t, restarted)

	require.Len(t, artifacts, 1)
	require.Equal(t, "rendered-page", artifacts[0].GetArtifactId())
	require.Equal(t, uint64(1), artifacts[0].GetRevision())
}
