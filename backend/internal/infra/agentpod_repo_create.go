package infra

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"gorm.io/gorm"
)

func createPod(tx *gorm.DB, pod *agentpod.Pod) error {
	if err := assignPodRunnerCluster(tx, pod); err != nil {
		return err
	}
	if err := tx.Create(pod).Error; err != nil {
		if isUniqueConstraintViolation(err, "idx_pods_source_pod_key_active_unique") {
			return agentpod.ErrSandboxAlreadyResumed
		}
		if isUniqueConstraintViolation(err, "idx_pods_orchestration_worker_launch") ||
			isUniqueConstraintViolation(
				err,
				"pods.organization_id, pods.orchestration_worker_launch_id",
			) {
			return agentpod.ErrWorkerLaunchPodAlreadyExists
		}
		return err
	}
	return nil
}

func assignPodRunnerCluster(tx *gorm.DB, pod *agentpod.Pod) error {
	if pod.ClusterID > 0 {
		return nil
	}
	var row struct {
		ClusterID int64
	}
	if err := tx.Table("runners").
		Select("cluster_id").
		Where("id = ? AND organization_id = ?", pod.RunnerID, pod.OrganizationID).
		Take(&row).Error; err != nil {
		return fmt.Errorf("resolve runner cluster for pod: %w", err)
	}
	if row.ClusterID <= 0 {
		return fmt.Errorf("runner %d has no execution cluster", pod.RunnerID)
	}
	pod.ClusterID = row.ClusterID
	return nil
}
