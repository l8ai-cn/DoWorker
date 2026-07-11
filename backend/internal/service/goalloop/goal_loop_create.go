package goalloop

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func (s *Service) Create(ctx context.Context, req CreateRequest) (*domain.GoalLoop, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}
	if err := s.validateWorkerSpecSnapshot(ctx, req.OrganizationID, req.WorkerSpecSnapshotID); err != nil {
		return nil, err
	}
	slug, err := s.ensureUniqueSlug(ctx, req.OrganizationID, req.Slug, req.Name)
	if err != nil {
		return nil, err
	}
	criteria, _ := json.Marshal(trimCriteria(req.AcceptanceCriteria))
	loop := &domain.GoalLoop{
		OrganizationID:       req.OrganizationID,
		CreatedByID:          req.CreatedByID,
		Name:                 strings.TrimSpace(req.Name),
		Slug:                 slug,
		Description:          trimOptional(req.Description),
		WorkerSpecSnapshotID: req.WorkerSpecSnapshotID,
		Objective:            strings.TrimSpace(req.Objective),
		AcceptanceCriteria:   criteria,
		VerificationCommand:  strings.TrimSpace(req.VerificationCommand),
		Status:               domain.StatusDraft,
		MaxIterations:        defaultInt(req.MaxIterations, 10),
		TokenBudget:          req.TokenBudget,
		TimeoutMinutes:       defaultInt(req.TimeoutMinutes, 60),
		NoProgressLimit:      defaultInt(req.NoProgressLimit, 3),
		SameErrorLimit:       defaultInt(req.SameErrorLimit, 2),
		EscalationPolicy:     defaultString(req.EscalationPolicy, domain.EscalationPause),
	}
	if err := s.repo.Create(ctx, loop); err != nil {
		return nil, err
	}
	return loop, nil
}

func (s *Service) validateWorkerSpecSnapshot(
	ctx context.Context,
	organizationID, snapshotID int64,
) error {
	if s.workerSpecs == nil {
		return ErrExecutionUnavailable
	}
	snapshot, err := s.workerSpecs.GetByID(ctx, organizationID, snapshotID)
	if errors.Is(err, workerspecdomain.ErrNotFound) {
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	if snapshot.OrganizationID != organizationID {
		return ErrInvalidInput
	}
	return nil
}

func (s *Service) ensureUniqueSlug(ctx context.Context, orgID int64, requested, name string) (string, error) {
	if requested != "" {
		if err := slugkit.Validate(requested); err != nil {
			return "", ErrInvalidInput
		}
		exists, err := s.repo.ExistsSlug(ctx, orgID, requested)
		if err != nil {
			return "", err
		}
		if exists {
			return "", ErrInvalidInput
		}
		return requested, nil
	}
	return slugkit.GenerateUnique(ctx, name, slugkit.FromExistsCheck(func(ctx context.Context, slug string) (bool, error) {
		return s.repo.ExistsSlug(ctx, orgID, slug)
	}))
}

func validateCreate(req CreateRequest) error {
	if req.OrganizationID <= 0 || req.CreatedByID <= 0 || req.WorkerSpecSnapshotID <= 0 {
		return ErrInvalidInput
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Objective) == "" ||
		strings.TrimSpace(req.VerificationCommand) == "" || len(trimCriteria(req.AcceptanceCriteria)) == 0 {
		return ErrInvalidInput
	}
	if req.MaxIterations != 0 && (req.MaxIterations < 1 || req.MaxIterations > 100) {
		return ErrInvalidInput
	}
	if req.TimeoutMinutes != 0 && (req.TimeoutMinutes < 1 || req.TimeoutMinutes > 1440) {
		return ErrInvalidInput
	}
	if req.NoProgressLimit != 0 && (req.NoProgressLimit < 1 || req.NoProgressLimit > 20) {
		return ErrInvalidInput
	}
	if req.SameErrorLimit != 0 && (req.SameErrorLimit < 1 || req.SameErrorLimit > 20) {
		return ErrInvalidInput
	}
	policy := defaultString(req.EscalationPolicy, domain.EscalationPause)
	if policy != domain.EscalationPause && policy != domain.EscalationFail {
		return ErrInvalidInput
	}
	if req.TokenBudget != nil && *req.TokenBudget <= 0 {
		return ErrInvalidInput
	}
	return nil
}

func trimCriteria(criteria []string) []string {
	out := make([]string, 0, len(criteria))
	for _, criterion := range criteria {
		if value := strings.TrimSpace(criterion); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func trimOptional(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	out := strings.TrimSpace(*value)
	return &out
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
