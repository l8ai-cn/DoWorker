package runner

import (
	"encoding/json"
	"fmt"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/workbench"
)

func (f *acpWorkbenchForwarder) artifactPublishStarted(
	executionID string,
	declaration json.RawMessage,
) {
	f.toolUpdate("", acp.ToolCallUpdate{
		ToolCallID:    executionID,
		ToolName:      "workbench.publish_artifact",
		Status:        "running",
		ArgumentsJSON: string(declaration),
	})
}

func (f *acpWorkbenchForwarder) artifactPublishFailed(
	executionID string,
	err error,
) {
	f.toolResult("", acp.ToolCallResult{
		ToolCallID:   executionID,
		ToolName:     "workbench.publish_artifact",
		Success:      false,
		ErrorMessage: err.Error(),
	})
}

func (f *acpWorkbenchForwarder) artifactPublished(
	executionID string,
	published *workbench.PublishedArtifactDeclaration,
) error {
	artifact, err := f.observer.PublishedArtifact(
		published.ArtifactID,
		executionID,
		f.currentCommandID(),
	)
	if err != nil {
		return err
	}
	result, err := json.Marshal(published)
	if err != nil {
		return fmt.Errorf("marshal artifact publication: %w", err)
	}
	f.toolResult("", acp.ToolCallResult{
		ToolCallID: executionID,
		ToolName:   "workbench.publish_artifact",
		Success:    true,
		ResultText: string(result),
	})
	f.send(f.mapper.Artifacts([]*workbench.ArtifactDescriptor{artifact}))
	return nil
}
