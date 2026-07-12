package runner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	runnerdomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

func (s *Service) AuthorizeRunner(ctx context.Context, authKey string, orgID, userID, clusterID int64, nodeID string) (*runnerdomain.Runner, error) {
	pendingAuth, err := s.repo.GetPendingAuthByKey(ctx, authKey)
	if err != nil {
		return nil, err
	}
	if pendingAuth == nil {
		return nil, ErrAuthRequestNotFound
	}
	if pendingAuth.IsExpired() {
		return nil, ErrAuthRequestExpired
	}
	if err := s.requireExecutionCluster(ctx, clusterID, orgID); err != nil {
		return nil, err
	}
	finalNodeID, err := requestedNodeID(pendingAuth, nodeID)
	if err != nil {
		return nil, err
	}
	if s.billingService != nil {
		if err := s.billingService.CheckQuota(ctx, orgID, "runners", 1); err != nil {
			return nil, ErrRunnerQuotaExceeded
		}
	}
	exists, err := s.repo.ExistsByNodeIDAndOrg(ctx, orgID, finalNodeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrRunnerAlreadyExists
	}

	r := &runnerdomain.Runner{
		OrganizationID:     orgID,
		ClusterID:          clusterID,
		NodeID:             finalNodeID,
		Status:             runnerdomain.RunnerStatusOffline,
		MaxConcurrentPods:  5,
		Visibility:         runnerdomain.VisibilityOrganization,
		RegisteredByUserID: &userID,
		Tags:               registrationLabelsToTags(pendingAuth.Labels),
	}
	rowsAffected, err := s.repo.AuthorizePendingAuthAtomic(ctx, pendingAuth.ID, orgID, clusterID, r)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize runner: %w", err)
	}
	if rowsAffected == 0 {
		return nil, ErrAuthRequestAlreadyAuthorized
	}
	return r, nil
}

func requestedNodeID(pendingAuth *runnerdomain.PendingAuth, nodeID string) (string, error) {
	if nodeID != "" {
		return nodeID, nil
	}
	if pendingAuth.NodeID != nil {
		return *pendingAuth.NodeID, nil
	}
	nodeIDBytes := make([]byte, 8)
	if _, err := rand.Read(nodeIDBytes); err != nil {
		return "", fmt.Errorf("failed to generate node ID: %w", err)
	}
	return fmt.Sprintf("runner-%s", hex.EncodeToString(nodeIDBytes)), nil
}
