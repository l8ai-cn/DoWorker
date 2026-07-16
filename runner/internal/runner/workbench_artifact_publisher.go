package runner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/agentsmesh/runner/internal/workbench"
)

func (r *Runner) PublishWorkbenchArtifact(
	ctx context.Context,
	podKey string,
	executionID string,
	declaration json.RawMessage,
) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pod, ok := r.podStore.Get(podKey)
	if !ok {
		return nil, fmt.Errorf("pod %q is not active", podKey)
	}
	root, err := podWorkspaceRoot(pod)
	if err != nil {
		return nil, err
	}
	forwarder := pod.workbenchForwarder
	if forwarder == nil {
		return nil, fmt.Errorf("pod %q has no Agent Workbench publisher", podKey)
	}
	forwarder.artifactMu.Lock()
	defer forwarder.artifactMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	forwarder.artifactPublishStarted(executionID, declaration)
	published, err := workbench.PublishArtifactDeclaration(root, declaration)
	if err != nil {
		forwarder.artifactPublishFailed(executionID, err)
		return nil, err
	}
	if err := forwarder.artifactPublished(executionID, published); err != nil {
		forwarder.artifactPublishFailed(executionID, err)
		return nil, err
	}
	return published, nil
}
