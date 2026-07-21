package agentworkbench

import (
	"fmt"
	"strings"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

func attachmentFileIDs(
	attachments []*agentworkbenchv2.ContentBlock,
) ([]string, error) {
	fileIDs := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment == nil || attachment.Identity == nil ||
			attachment.Identity.Namespace != "agentcloud.session-file" ||
			attachment.Identity.SemanticKey != "attachment" ||
			attachment.Identity.SchemaVersion != "1" {
			return nil, ErrInvalidCommand
		}
		file := attachment.GetFile()
		if file == nil || strings.TrimSpace(file.ArtifactId) == "" {
			return nil, ErrInvalidCommand
		}
		fileIDs = append(fileIDs, strings.TrimSpace(file.ArtifactId))
	}
	return fileIDs, nil
}

func materializedPrompt(text string, attachmentPaths []string) string {
	parts := make([]string, 0, len(attachmentPaths)+1)
	for _, attachmentPath := range attachmentPaths {
		parts = append(parts, fmt.Sprintf("[Attached: %s]", attachmentPath))
	}
	if normalized := strings.TrimSpace(text); normalized != "" {
		if len(parts) > 0 {
			parts = append(parts, "")
		}
		parts = append(parts, normalized)
	}
	return strings.Join(parts, "\n")
}

func promptItemContent(text string, attachmentPaths []string) []map[string]string {
	content := make([]map[string]string, 0, len(attachmentPaths)+1)
	if normalized := strings.TrimSpace(text); normalized != "" {
		content = append(content, map[string]string{
			"type": "input_text",
			"text": normalized,
		})
	}
	for _, attachmentPath := range attachmentPaths {
		content = append(content, map[string]string{
			"type": "input_file",
			"path": attachmentPath,
		})
	}
	return content
}
