package infra

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"gorm.io/gorm"
)

func validateRunnerCluster(tx *gorm.DB, rn *runner.Runner) error {
	if rn.ClusterID <= 0 {
		return fmt.Errorf("runner execution cluster is required")
	}
	return validateExecutionCluster(tx, rn.OrganizationID, rn.ClusterID)
}

func validateExecutionCluster(tx *gorm.DB, organizationID, clusterID int64) error {
	var count int64
	if err := tx.Table("execution_clusters").
		Where("id = ? AND organization_id = ?", clusterID, organizationID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("validate execution cluster: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("execution cluster %d does not belong to organization %d", clusterID, organizationID)
	}
	return nil
}
