package workbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestOfficePreviewDescriptorLifecycle(t *testing.T) {
	artifact := readyArtifactDescriptor(artifactFile{
		path:      "outputs/report.docx",
		filename:  "report.docx",
		mediaType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		digest:    "sha256:source",
		byteSize:  42,
	}, 3)
	source, ok := ResolveOfficePreviewSource(artifact)
	require.True(t, ok)
	require.Equal(t, "outputs/report.docx", source.Path)
	require.Equal(t, "report.pdf", source.Filename)

	processing := OfficePreviewProcessing(artifact, source)
	require.Len(t, processing.GetRepresentations(), 2)
	require.Equal(t,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_PROCESSING,
		officePreviewRepresentation(processing).GetStatus(),
	)
	require.Equal(t,
		[]string{"original", OfficePreviewRepresentationID},
		processing.GetRevisions()[0].GetRepresentationIds(),
	)
	require.Equal(t,
		[]string{"original", OfficePreviewRepresentationID},
		processing.GetGrants()[0].GetRepresentationIds(),
	)

	ready := OfficePreviewReady(
		processing,
		"workspace:.agent-cloud/workbench/previews/preview-abcd.pdf",
		"sha256:pdf",
		128,
	)
	preview := officePreviewRepresentation(ready)
	require.Equal(t,
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		preview.GetStatus(),
	)
	require.Equal(t, "primary", preview.GetRole())
	require.Equal(t, "workspace:.agent-cloud/workbench/previews/preview-abcd.pdf",
		preview.GetTransport().GetResourceId())
	require.Equal(t, uint64(128), preview.GetByteSize())
}
