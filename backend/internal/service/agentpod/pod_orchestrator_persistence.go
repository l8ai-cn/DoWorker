package agentpod

func newPodServiceCreateRequest(
	req *OrchestrateCreatePodRequest,
	resolved *agentfileResolved,
	repositoryID *int64,
	branchName *string,
	sessionID string,
	interactionMode string,
	model string,
	permissionMode string,
	initialStatus string,
) *CreatePodRequest {
	agentfileLayer := ""
	if req.AgentfileLayer != nil {
		agentfileLayer = *req.AgentfileLayer
	}
	return &CreatePodRequest{
		OrganizationID:                 req.OrganizationID,
		RunnerID:                       req.RunnerID,
		ClusterID:                      req.clusterID,
		AgentSlug:                      req.AgentSlug,
		RepositoryID:                   repositoryID,
		TicketID:                       req.TicketID,
		CreatedByID:                    req.UserID,
		Prompt:                         resolved.Prompt,
		Alias:                          req.Alias,
		BranchName:                     branchName,
		Model:                          model,
		PermissionMode:                 permissionMode,
		SessionID:                      sessionID,
		SourcePodKey:                   req.SourcePodKey,
		InteractionMode:                interactionMode,
		AutomationLevel:                req.AutomationLevel,
		Perpetual:                      req.Perpetual,
		ResolvedConfig:                 resolved.ConfigValues,
		InitialStatus:                  initialStatus,
		ModelResourceID:                req.ModelResourceID,
		AgentfileLayer:                 agentfileLayer,
		ResolvedWorkerSpec:             req.resolvedWorkerSpec,
		WorkerDependencyArtifactJSON:   workerDependencyArtifactJSON(req),
		WorkerDependencyArtifactDigest: workerDependencyArtifactDigest(req),
		WorkerSpecSnapshotID:           req.workerSpecSnapshotID,
		OrchestrationWorkerLaunchID:    req.OrchestrationWorkerLaunchID,
	}
}

func workerDependencyArtifactJSON(req *OrchestrateCreatePodRequest) []byte {
	if req == nil || req.preResolvedArtifact == nil {
		return append([]byte(nil), req.preResolvedArtifactJSON...)
	}
	return req.preResolvedArtifact.JSON()
}

func workerDependencyArtifactDigest(req *OrchestrateCreatePodRequest) string {
	if req == nil || req.preResolvedArtifact == nil {
		return req.preResolvedArtifactDigest
	}
	return req.preResolvedArtifact.Digest()
}
