package runner

import (
	"context"
	"fmt"
	"sort"

	runnerdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
)

func (s *Service) requireExecutionCluster(ctx context.Context, clusterID, organizationID int64) error {
	if clusterID <= 0 || s.clusterRepo == nil {
		return ErrExecutionClusterNotFound
	}
	cluster, err := s.clusterRepo.GetByIDAndOrganization(ctx, clusterID, organizationID)
	if err != nil {
		return fmt.Errorf("get execution cluster: %w", err)
	}
	if cluster == nil {
		return ErrExecutionClusterNotFound
	}
	return nil
}

func registrationLabelsToTags(labels runnerdomain.Labels) runnerdomain.StringSlice {
	if len(labels) == 0 {
		return nil
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	tags := make(runnerdomain.StringSlice, 0, len(keys))
	for _, key := range keys {
		tags = append(tags, key+"="+labels[key])
	}
	return tags
}
