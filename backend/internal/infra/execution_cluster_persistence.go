package infra

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"gorm.io/gorm"
)

func localExecutionClusterID(tx *gorm.DB, organizationID int64) (int64, error) {
	var row struct {
		ID int64
	}
	if err := tx.Table("execution_clusters").
		Select("id").
		Where("organization_id = ? AND slug = ?", organizationID, "local").
		Take(&row).Error; err != nil {
		return 0, fmt.Errorf("resolve local execution cluster: %w", err)
	}
	return row.ID, nil
}

func assignLocalClusterToToken(tx *gorm.DB, token *runner.GRPCRegistrationToken) error {
	if token.ClusterID == 0 {
		clusterID, err := localExecutionClusterID(tx, token.OrganizationID)
		if err != nil {
			return err
		}
		token.ClusterID = clusterID
	}
	return validateExecutionCluster(tx, token.OrganizationID, token.ClusterID)
}

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
