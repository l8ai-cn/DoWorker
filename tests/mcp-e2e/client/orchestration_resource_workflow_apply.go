package client

import "context"

type workflowApplyResponseWire struct {
	Resource             resourceApplyWire `json:"resource"`
	WorkflowID           string            `json:"workflowId"`
	WorkerSpecSnapshotID string            `json:"workerSpecSnapshotId"`
}

func (r *REST) applyWorkflowResource(
	ctx context.Context,
	request map[string]string,
) (AppliedOrchestrationResource, error) {
	var response workflowApplyResponseWire
	err := r.connectCall(
		ctx,
		orchestrationResourceService+"ApplyWorkflowPlan",
		request,
		&response,
	)
	applied, applyErr := response.Resource.applied(0, err)
	if applyErr != nil {
		return AppliedOrchestrationResource{}, applyErr
	}
	workflowID, parseErr := parsePositiveID(
		"workflow",
		response.WorkflowID,
	)
	if parseErr != nil {
		return AppliedOrchestrationResource{}, parseErr
	}
	snapshotID, parseErr := parsePositiveID(
		"worker spec snapshot",
		response.WorkerSpecSnapshotID,
	)
	if parseErr != nil {
		return AppliedOrchestrationResource{}, parseErr
	}
	applied.WorkflowID = workflowID
	applied.WorkerSpecSnapshotID = snapshotID
	return applied, nil
}
