package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestAttachmentFileIDsAcceptsSessionFileReference(t *testing.T) {
	ids, err := attachmentFileIDs([]*agentworkbenchv2.ContentBlock{{
		ContentId: "file-1",
		Identity: &agentworkbenchv2.ContentIdentity{
			Namespace: "agentcloud.session-file", SemanticKey: "attachment",
			SchemaVersion: "1",
		},
		Content: &agentworkbenchv2.ContentBlock_File{
			File: &agentworkbenchv2.MediaContent{ArtifactId: "file-1"},
		},
	}})

	require.NoError(t, err)
	require.Equal(t, []string{"file-1"}, ids)
	require.Equal(
		t,
		"[Attached: uploads/file-1-brief.txt]\n\nReview this file",
		materializedPrompt("Review this file", []string{"uploads/file-1-brief.txt"}),
	)
}

func TestAttachmentFileIDsRejectsForeignIdentity(t *testing.T) {
	_, err := attachmentFileIDs([]*agentworkbenchv2.ContentBlock{{
		Identity: &agentworkbenchv2.ContentIdentity{
			Namespace: "foreign", SemanticKey: "attachment", SchemaVersion: "1",
		},
	}})

	require.ErrorIs(t, err, ErrInvalidCommand)
}
