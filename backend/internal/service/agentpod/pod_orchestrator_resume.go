package agentpod

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	userService "github.com/anthropics/agentsmesh/backend/internal/service/user"
)

func (o *PodOrchestrator) handleResumeMode(ctx context.Context, req *OrchestrateCreatePodRequest) (*podDomain.Pod, string, error) {
	sourcePod, err := o.podService.GetPod(ctx, req.SourcePodKey)
	if err != nil {
		return nil, "", ErrSourcePodNotFound
	}

	if sourcePod.OrganizationID != req.OrganizationID {
		return nil, "", ErrSourcePodAccessDenied
	}

	if sourcePod.Status != podDomain.StatusTerminated &&
		sourcePod.Status != podDomain.StatusCompleted &&
		sourcePod.Status != podDomain.StatusOrphaned {
		return nil, "", ErrSourcePodNotTerminated
	}

	existingResumePod, err := o.podService.GetActivePodBySourcePodKey(ctx, req.SourcePodKey)
	if err == nil && existingResumePod != nil {
		return nil, "", ErrSourcePodAlreadyResumed
	}

	if req.RunnerID == 0 {
		req.RunnerID = sourcePod.RunnerID
	} else if sourcePod.RunnerID != req.RunnerID {
		return nil, "", ErrResumeRunnerMismatch
	}
	req.clusterID = sourcePod.ClusterID

	if req.AgentSlug != "" && req.AgentSlug != sourcePod.AgentSlug {
		return nil, "", ErrResumeAgentMismatch
	}
	if req.AgentSlug == "" {
		req.AgentSlug = sourcePod.AgentSlug
	}
	if err := validateResumeWorkerInput(req, sourcePod); err != nil {
		return nil, "", err
	}
	if req.RepositoryID == nil {
		req.RepositoryID = sourcePod.RepositoryID
	}
	if req.TicketID == nil {
		req.TicketID = sourcePod.TicketID
	}
	if req.BranchName == nil {
		req.BranchName = sourcePod.BranchName
	}
	if req.ModelResourceID == nil {
		req.ModelResourceID = sourcePod.ModelResourceID
	}
	if err := o.inheritWorkerSpecSnapshot(ctx, req, sourcePod); err != nil {
		return nil, "", err
	}
	if err := o.inheritResumeState(ctx, req, sourcePod); err != nil {
		return nil, "", err
	}
	appendWorkerSpecPromptOverride(req)
	req.Perpetual = sourcePod.Perpetual

	var sessionID string
	if sourcePod.SessionID != nil && *sourcePod.SessionID != "" {
		sessionID = *sourcePod.SessionID
	} else {
		sessionID = uuid.New().String()
	}

	return sourcePod, sessionID, nil
}

func validateResumeWorkerInput(
	req *OrchestrateCreatePodRequest,
	sourcePod *podDomain.Pod,
) error {
	if sourcePod.WorkerSpecSnapshotID == nil ||
		*sourcePod.WorkerSpecSnapshotID <= 0 {
		return ErrWorkerSpecSnapshotUnavailable
	}
	if req.RepositoryID != nil &&
		!int64PointersEqual(req.RepositoryID, sourcePod.RepositoryID) {
		return ErrWorkerSpecSnapshotMismatch
	}
	if req.ModelResourceID != nil &&
		!int64PointersEqual(req.ModelResourceID, sourcePod.ModelResourceID) {
		return ErrWorkerSpecSnapshotMismatch
	}
	if req.BranchName != nil &&
		!stringPointerMatches(req.BranchName, workerSpecStringValue(sourcePod.BranchName)) {
		return ErrWorkerSpecSnapshotMismatch
	}
	if req.AutomationLevel != "" &&
		req.AutomationLevel != sourcePod.AutomationLevel {
		return ErrWorkerSpecSnapshotMismatch
	}
	if req.TokenBudget != nil ||
		len(req.KnowledgeMounts) > 0 ||
		len(req.ModelResourceEnv) > 0 ||
		len(req.ModelResourceArgs) > 0 {
		return ErrWorkerSpecSnapshotMismatch
	}
	return nil
}

// getUserGitCredential retrieves the default Git credential for a user.
// Returns nil if using runner_local (Runner will use local Git config).
func (o *PodOrchestrator) getUserGitCredential(ctx context.Context, userID int64) *userService.DecryptedCredential {
	if o.userService == nil {
		return nil
	}

	defaultCred, err := o.userService.GetDefaultGitCredential(ctx, userID)
	if err != nil || defaultCred == nil {
		return nil
	}

	if defaultCred.CredentialType == "runner_local" {
		return nil
	}

	decrypted, err := o.userService.GetDecryptedCredentialToken(ctx, userID, defaultCred.ID)
	if err != nil {
		slog.WarnContext(ctx, "failed to decrypt Git credential", "user_id", userID, "error", err)
		return nil
	}

	return decrypted
}
