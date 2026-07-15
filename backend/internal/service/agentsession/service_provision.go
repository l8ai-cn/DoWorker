package agentsession

import (
	"context"
	"errors"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"gorm.io/gorm"
)

var ErrPodIdentityMismatch = errors.New("session and pod identities do not match")
var ErrSessionBindingChanged = errors.New("session pod binding changed")

func (s *Service) PrepareForPod(
	ctx context.Context,
	pod *podDomain.Pod,
	spec domain.ProvisionSpec,
) (*domain.ProvisionReceipt, error) {
	if pod == nil {
		return nil, errors.New("pod is required")
	}
	if spec.UpdateExisting {
		return s.rebindSession(ctx, pod, spec)
	}
	id := spec.ID
	if id == "" {
		var err error
		id, err = NewID()
		if err != nil {
			return nil, err
		}
	}
	title := spec.Title
	if title == nil {
		title = pod.Alias
	}
	if title == nil {
		title = pod.Title
	}
	row := &domain.Session{
		ID:              id,
		OrganizationID:  pod.OrganizationID,
		UserID:          pod.CreatedByID,
		PodKey:          pod.PodKey,
		AgentSlug:       pod.AgentSlug,
		Title:           title,
		ParentSessionID: spec.ParentSessionID,
		Status:          "idle",
	}
	if err := s.Create(ctx, row); err != nil {
		return nil, err
	}
	return &domain.ProvisionReceipt{Session: row, Created: true}, nil
}

func (s *Service) rebindSession(
	ctx context.Context,
	pod *podDomain.Pod,
	spec domain.ProvisionSpec,
) (*domain.ProvisionReceipt, error) {
	if spec.ID == "" {
		return nil, ErrNotFound
	}
	if spec.ExpectedPodKey == "" {
		return nil, ErrSessionBindingChanged
	}
	var row domain.Session
	var previousAgentSlug string
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND deleted_at IS NULL", spec.ID).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if row.OrganizationID != pod.OrganizationID || row.UserID != pod.CreatedByID {
			return ErrPodIdentityMismatch
		}
		previousAgentSlug = row.AgentSlug
		now := time.Now()
		result := tx.Model(&domain.Session{}).
			Where(
				"id = ? AND organization_id = ? AND user_id = ? AND pod_key = ? AND deleted_at IS NULL",
				spec.ID,
				pod.OrganizationID,
				pod.CreatedByID,
				spec.ExpectedPodKey,
			).
			Updates(map[string]any{
				"agent_slug": pod.AgentSlug,
				"pod_key":    pod.PodKey,
				"updated_at": now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrSessionBindingChanged
		}
		row.AgentSlug = pod.AgentSlug
		row.PodKey = pod.PodKey
		row.UpdatedAt = now
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &domain.ProvisionReceipt{
		Session:           &row,
		PreviousPodKey:    spec.ExpectedPodKey,
		PreviousAgentSlug: previousAgentSlug,
	}, nil
}

func (s *Service) RollbackProvision(
	ctx context.Context,
	receipt *domain.ProvisionReceipt,
) error {
	if receipt == nil || receipt.Session == nil {
		return errors.New("provision receipt is required")
	}
	if receipt.Created {
		return s.rollbackCreatedSession(ctx, receipt.Session)
	}
	return s.rollbackReboundSession(ctx, receipt)
}

func (s *Service) rollbackCreatedSession(ctx context.Context, row *domain.Session) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM conversation_items WHERE session_id = ?",
			row.ID,
		).Error; err != nil {
			return err
		}
		result := tx.Unscoped().
			Where("id = ? AND pod_key = ?", row.ID, row.PodKey).
			Delete(&domain.Session{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrSessionBindingChanged
		}
		return nil
	})
}

func (s *Service) rollbackReboundSession(
	ctx context.Context,
	receipt *domain.ProvisionReceipt,
) error {
	result := s.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where(
			"id = ? AND organization_id = ? AND user_id = ? AND pod_key = ? AND deleted_at IS NULL",
			receipt.Session.ID,
			receipt.Session.OrganizationID,
			receipt.Session.UserID,
			receipt.Session.PodKey,
		).
		Updates(map[string]any{
			"agent_slug": receipt.PreviousAgentSlug,
			"pod_key":    receipt.PreviousPodKey,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrSessionBindingChanged
	}
	return nil
}
