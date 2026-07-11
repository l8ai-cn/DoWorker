package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *podRepo) UpdatePreviewConfig(
	ctx context.Context,
	podKey string,
	createdByID int64,
	previewPort int,
	previewPath string,
) (*agentpod.Pod, error) {
	var updated *agentpod.Pod
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pod, err := lockPodWithActiveConfig(tx, podKey)
		if err != nil {
			return err
		}
		revision, err := nextPreviewConfigRevision(tx, pod, createdByID, previewPort, previewPath)
		if err != nil {
			return err
		}
		result := tx.Model(&agentpod.PodConfigRevision{}).
			Where("id = ? AND status = ?", pod.ActiveConfigRevision.ID, agentpod.ConfigRevisionStatusActive).
			Update("status", agentpod.ConfigRevisionStatusSuperseded)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("active config revision changed while updating preview config")
		}
		if err := tx.Create(revision).Error; err != nil {
			return err
		}
		if err := tx.Model(&agentpod.Pod{}).Where("id = ?", pod.ID).Updates(map[string]interface{}{
			"preview_port":              previewPort,
			"preview_path":              previewPath,
			"generation":                revision.Revision,
			"active_config_revision_id": revision.ID,
		}).Error; err != nil {
			return err
		}
		var persisted agentpod.Pod
		if err := tx.Preload("ActiveConfigRevision").First(&persisted, pod.ID).Error; err != nil {
			return err
		}
		updated = &persisted
		return nil
	})
	return updated, err
}

func lockPodWithActiveConfig(tx *gorm.DB, podKey string) (*agentpod.Pod, error) {
	var pod agentpod.Pod
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("ActiveConfigRevision").
		Where("pod_key = ?", podKey).
		First(&pod).Error
	if err != nil {
		return nil, err
	}
	if pod.ActiveConfigRevision == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &pod, nil
}

func nextPreviewConfigRevision(
	tx *gorm.DB,
	pod *agentpod.Pod,
	createdByID int64,
	previewPort int,
	previewPath string,
) (*agentpod.PodConfigRevision, error) {
	var nextRevision int64
	if err := tx.Model(&agentpod.PodConfigRevision{}).
		Where("pod_id = ?", pod.ID).
		Select("COALESCE(MAX(revision), 0) + 1").
		Scan(&nextRevision).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	active := pod.ActiveConfigRevision
	return &agentpod.PodConfigRevision{
		PodID:           pod.ID,
		Revision:        nextRevision,
		AgentfileLayer:  active.AgentfileLayer,
		Status:          agentpod.ConfigRevisionStatusActive,
		ConfigSummary:   append([]byte(nil), active.ConfigSummary...),
		ModelResourceID: active.ModelResourceID,
		PreviewPort:     previewPort,
		PreviewPath:     previewPath,
		CreatedByID:     createdByID,
		AppliedAt:       &now,
	}, nil
}
