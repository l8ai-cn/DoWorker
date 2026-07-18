package runner

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/stretchr/testify/require"
)

func TestACPWorkbenchForwarderEmitsArtifactsWhenTurnBecomesIdle(t *testing.T) {
	root := t.TempDir()
	connection := client.NewMockConnection()
	forwarder, err := newACPWorkbenchForwarder(
		"pod-1",
		"codex",
		root,
		connection,
	)
	require.NoError(t, err)
	output := filepath.Join(root, "output", "demo.mp4")
	require.NoError(t, os.MkdirAll(filepath.Dir(output), 0o755))
	require.NoError(t, os.WriteFile(output, []byte("video"), 0o644))

	forwarder.state(acp.StateIdle)

	raw := rawWorkbenchMessages(t, connection.GetEvents())
	require.Len(t, raw, 2)
	require.NotNil(t, raw[0].GetWorkbenchEvents().GetMutations()[0].GetStatus())
	artifact := raw[1].GetWorkbenchEvents().GetMutations()[0].GetArtifact()
	require.Equal(t, "workspace:output/demo.mp4", artifact.GetArtifactId())
}

func TestACPWorkbenchForwarderPublishesOfficePDFPreview(t *testing.T) {
	root := t.TempDir()
	connection := client.NewMockConnection()
	forwarder, err := newACPWorkbenchForwarder(
		"pod-1",
		"codex",
		root,
		connection,
	)
	require.NoError(t, err)
	forwarder.convertOffice = func(
		_ context.Context,
		_, _ string,
	) ([]byte, error) {
		return []byte("%PDF-preview"), nil
	}
	output := filepath.Join(root, "outputs", "report.docx")
	require.NoError(t, os.MkdirAll(filepath.Dir(output), 0o755))
	require.NoError(t, os.WriteFile(output, []byte("office-source"), 0o644))

	forwarder.state(acp.StateIdle)

	require.Eventually(t, func() bool {
		return len(rawWorkbenchMessages(t, connection.GetEvents())) >= 4
	}, time.Second, 10*time.Millisecond)
	raw := rawWorkbenchMessages(t, connection.GetEvents())
	processing := raw[2].GetWorkbenchEvents().GetMutations()[0].GetArtifact()
	ready := raw[3].GetWorkbenchEvents().GetMutations()[0].GetArtifact()
	require.Equal(t,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING,
		processing.GetRepresentations()[1].GetStatus(),
	)
	preview := ready.GetRepresentations()[1]
	require.Equal(t,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		preview.GetStatus(),
	)
	resourceID := preview.GetTransport().GetResourceId()
	require.Contains(t, resourceID, "workspace:.do-worker/workbench/previews/")
	previewPath := filepath.Join(
		root,
		filepath.FromSlash(resourceID[len("workspace:"):]),
	)
	content, err := os.ReadFile(previewPath)
	require.NoError(t, err)
	require.Equal(t, []byte("%PDF-preview"), content)
}

func TestACPWorkbenchForwarderDoesNotGuessArtifactToolFromResult(t *testing.T) {
	root := t.TempDir()
	connection := client.NewMockConnection()
	forwarder, err := newACPWorkbenchForwarder(
		"pod-1",
		"codex",
		root,
		connection,
	)
	require.NoError(t, err)
	forwarder.toolUpdate("", acp.ToolCallUpdate{
		ToolCallID: "unrelated-tool",
		ToolName:   "shell",
		Status:     "running",
	})
	connection.Reset()
	writePublishedVideoArtifact(t, root)

	forwarder.toolResult("", acp.ToolCallResult{
		ToolCallID: "unrelated-tool",
		ToolName:   "shell",
		Success:    true,
	})

	raw := rawWorkbenchMessages(t, connection.GetEvents())
	require.Len(t, raw, 1)
	require.Nil(t, raw[0].GetWorkbenchEvents().GetMutations()[0].GetArtifact())
}

func TestRunnerPublishWorkbenchArtifactEmitsExactToolAndRevisionProvenance(
	t *testing.T,
) {
	root := t.TempDir()
	connection := client.NewMockConnection()
	forwarder, err := newACPWorkbenchForwarder(
		"pod-1",
		"codex",
		root,
		connection,
	)
	require.NoError(t, err)
	store := NewInMemoryPodStore()
	store.Put("pod-1", &Pod{
		PodKey:             "pod-1",
		WorkDir:            root,
		workbenchForwarder: forwarder,
	})
	runner := &Runner{podStore: store}
	forwarder.setActiveCommandID("command-1")
	writeVideoRepresentation(t, root, "video-one")

	_, err = runner.PublishWorkbenchArtifact(
		context.Background(),
		"pod-1",
		"publish-execution-1",
		[]byte(publishedVideoDeclaration(1)),
	)

	require.NoError(t, err)
	raw := rawWorkbenchMessages(t, connection.GetEvents())
	require.Len(t, raw, 3)
	started := raw[0].GetWorkbenchEvents().GetMutations()[0].
		GetTimeline().GetContent().GetToolExecution()
	require.Equal(t, "agentsmesh.runner", started.GetIdentity().GetNamespace())
	require.Equal(t, "artifact.publish", started.GetIdentity().GetSemanticKey())
	artifactBatch := raw[2].GetWorkbenchEvents()
	require.Len(t, artifactBatch.GetMutations(), 2)
	artifact := artifactBatch.GetMutations()[0].GetArtifact()
	require.Empty(t, artifact.GetProvenance().GetToolExecutionId())
	require.Equal(t,
		"publish-execution-1",
		artifact.GetRevisions()[0].GetProvenance().GetToolExecutionId(),
	)
	require.Equal(t,
		"command-1",
		artifact.GetRevisions()[0].GetProvenance().GetCommandId(),
	)
	tool := artifactBatch.GetMutations()[1].GetTimeline().GetContent().GetToolExecution()
	require.Equal(t, "publish-execution-1", tool.GetExecutionId())
	require.Equal(t, "demo-video", tool.GetResults()[0].GetArtifacts()[0].GetArtifactId())
}

func TestRunnerPublishWorkbenchArtifactUsesNewToolForNextRevision(t *testing.T) {
	root := t.TempDir()
	connection := client.NewMockConnection()
	forwarder, err := newACPWorkbenchForwarder(
		"pod-1",
		"codex",
		root,
		connection,
	)
	require.NoError(t, err)
	store := NewInMemoryPodStore()
	store.Put("pod-1", &Pod{
		PodKey:             "pod-1",
		WorkDir:            root,
		workbenchForwarder: forwarder,
	})
	runner := &Runner{podStore: store}
	forwarder.setActiveCommandID("command-1")
	writeVideoRepresentation(t, root, "video-one")
	_, err = runner.PublishWorkbenchArtifact(
		context.Background(),
		"pod-1",
		"publish-execution-1",
		[]byte(publishedVideoDeclaration(1)),
	)
	require.NoError(t, err)
	connection.Reset()
	forwarder.setActiveCommandID("command-2")
	writeVideoRepresentation(t, root, "video-two")

	_, err = runner.PublishWorkbenchArtifact(
		context.Background(),
		"pod-1",
		"publish-execution-2",
		[]byte(publishedVideoDeclaration(2)),
	)

	require.NoError(t, err)
	raw := rawWorkbenchMessages(t, connection.GetEvents())
	require.Len(t, raw, 3)
	artifactBatch := raw[2].GetWorkbenchEvents()
	require.Len(t, artifactBatch.GetMutations(), 2)
	artifact := artifactBatch.GetMutations()[0].GetArtifact()
	require.Equal(t, uint64(2), artifact.GetRevision())
	require.Equal(t,
		"publish-execution-2",
		artifact.GetRevisions()[0].GetProvenance().GetToolExecutionId(),
	)
	require.Equal(t,
		"command-2",
		artifact.GetRevisions()[0].GetProvenance().GetCommandId(),
	)
	tool := artifactBatch.GetMutations()[1].GetTimeline().GetContent().GetToolExecution()
	require.Equal(t, "publish-execution-2", tool.GetExecutionId())
	require.Equal(t, uint64(2), tool.GetResults()[0].GetArtifacts()[0].GetRevision())
}

func writePublishedVideoArtifact(t *testing.T, root string) {
	writePublishedVideoArtifactRevision(t, root, 1, "video")
}

func writePublishedVideoArtifactRevision(
	t *testing.T,
	root string,
	revision uint64,
	content string,
) {
	t.Helper()
	writeVideoRepresentation(t, root, content)
	declarationDirectory := filepath.Join(
		root,
		filepath.FromSlash(".do-worker/workbench/artifacts"),
	)
	require.NoError(t, os.MkdirAll(declarationDirectory, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(declarationDirectory, "demo-video.json"),
		[]byte(publishedVideoDeclaration(revision)),
		0o644,
	))
}

func writeVideoRepresentation(t *testing.T, root string, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "output"), 0o755))
	sum := sha1.Sum([]byte(content))
	color := "0x" + hex.EncodeToString(sum[:3])
	output, err := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c="+color+":s=16x16:d=0.2:r=5",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		filepath.Join(root, "output", "demo.mp4"),
	).CombinedOutput()
	require.NoError(t, err, string(output))
}

func publishedVideoDeclaration(revision uint64) string {
	return fmt.Sprintf(`{
			"schema_version":"agentsmesh.agent-workbench.artifact/v1",
			"artifact_id":"demo-video",
			"revision":%d,
			"role":"preview",
			"primary_representation_id":"playable",
			"producer":{"namespace":"agentsmesh.mcp","type":"artifact.publish"},
			"representations":[{
				"representation_id":"playable",
				"path":"output/demo.mp4",
				"media_type":"video/mp4"
			}],
			"manifest":{
				"kind":"video",
				"stage":"ready",
				"playable_representation_id":"playable"
			}
		}`, revision)
}

func rawWorkbenchMessages(
	t *testing.T,
	events []client.EventCall,
) []*runnerv1.RunnerMessage {
	t.Helper()
	messages := make([]*runnerv1.RunnerMessage, 0, len(events))
	for _, event := range events {
		if event.Type != "raw_message" {
			continue
		}
		message, ok := event.Data.(*runnerv1.RunnerMessage)
		require.True(t, ok)
		if message.GetWorkbenchEvents() != nil {
			messages = append(messages, message)
		}
	}
	return messages
}
