package infra

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	workerspecservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"gorm.io/gorm"
)

func (r *podRepo) CreateWithConfigAndWorkerSpec(
	ctx context.Context,
	pod *agentpod.Pod,
	revision *agentpod.PodConfigRevision,
	resolved workerspecservice.ResolvedSnapshot,
	artifactJSON []byte,
	artifactDigest string,
) error {
	if pod.OrganizationID != resolved.OrganizationID() {
		return fmt.Errorf("pod organization does not match workerspec snapshot")
	}
	if pod.WorkerSpecSnapshotID != nil {
		return fmt.Errorf("pod already has a workerspec snapshot")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		snapshot, err := createWorkerSpecSnapshot(tx, resolved)
		if err != nil {
			return err
		}
		if err := createWorkerSpecDependencyArtifact(
			tx,
			pod.OrganizationID,
			snapshot.ID,
			artifactJSON,
			artifactDigest,
		); err != nil {
			return err
		}
		pod.WorkerSpecSnapshotID = &snapshot.ID
		return createPodWithConfig(tx, pod, revision)
	})
}

func createPodWithConfig(
	tx *gorm.DB,
	pod *agentpod.Pod,
	revision *agentpod.PodConfigRevision,
) error {
	if err := createPod(tx, pod); err != nil {
		return err
	}
	revision.PodID = pod.ID
	if err := tx.Create(revision).Error; err != nil {
		return err
	}
	nextGeneration := pod.Generation + 1
	if err := tx.Model(pod).Updates(map[string]interface{}{
		"active_config_revision_id": revision.ID,
		"generation":                nextGeneration,
	}).Error; err != nil {
		return err
	}
	pod.Generation = nextGeneration
	pod.ActiveConfigRevisionID = &revision.ID
	pod.ActiveConfigRevision = revision
	return nil
}
