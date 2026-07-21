package expert

import (
	"context"
	"errors"
	"fmt"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	skilldomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

var (
	ErrPodAccessDenied               = errors.New("pod access denied")
	ErrPodWorkerSpecSnapshotRequired = errors.New("pod must have a workerspec snapshot")
	ErrWorkerSpecSnapshotUnavailable = errors.New("workerspec snapshot is unavailable")
	ErrWorkerSpecSnapshotMismatch    = errors.New("workerspec snapshot does not match the source pod")
	ErrWorkerSpecSkillUnavailable    = errors.New("workerspec skill catalog is unavailable")
)

type PublishFromPodRequest struct {
	OrganizationID int64
	UserID         int64
	PodKey         string
	Name           string
	Slug           string
	Description    *string
}

func (s *Service) PublishFromPod(ctx context.Context, req *PublishFromPodRequest) (*expertdom.Expert, error) {
	if s.pods == nil {
		return nil, errors.New("pod loader not configured")
	}
	pod, err := s.pods.GetPod(ctx, req.PodKey)
	if err != nil {
		return nil, err
	}
	if pod.OrganizationID != req.OrganizationID {
		return nil, ErrPodAccessDenied
	}
	if pod.WorkerSpecSnapshotID == nil {
		return nil, ErrPodWorkerSpecSnapshotRequired
	}
	if s.workerSpecs == nil {
		return nil, ErrWorkerSpecSnapshotUnavailable
	}
	snapshotID := *pod.WorkerSpecSnapshotID
	if snapshotID <= 0 {
		return nil, ErrWorkerSpecSnapshotMismatch
	}
	snapshot, err := s.workerSpecs.GetByID(
		ctx,
		req.OrganizationID,
		snapshotID,
	)
	if err != nil {
		if errors.Is(err, specdomain.ErrNotFound) {
			return nil, ErrWorkerSpecSnapshotMismatch
		}
		return nil, errors.Join(ErrWorkerSpecSnapshotUnavailable, err)
	}
	spec, err := specdomain.NormalizeAndValidate(snapshot.Spec)
	if err != nil {
		return nil, errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	if snapshot.ID != snapshotID ||
		snapshot.OrganizationID != req.OrganizationID {
		return nil, ErrWorkerSpecSnapshotMismatch
	}
	skillSlugs, err := s.resolveSnapshotSkillSlugs(
		ctx,
		req.OrganizationID,
		spec.Workspace.SkillIDs,
	)
	if err != nil {
		return nil, err
	}
	name := req.Name
	if name == "" && pod.Alias != nil {
		name = *pod.Alias
	}
	if name == "" {
		name = pod.PodKey
	}
	createReq := &CreateExpertRequest{
		OrganizationID:       req.OrganizationID,
		UserID:               req.UserID,
		Name:                 name,
		Slug:                 req.Slug,
		Description:          req.Description,
		AgentSlug:            spec.Runtime.WorkerType.Slug.String(),
		InteractionMode:      string(spec.TypeConfig.InteractionMode),
		AutomationLevel:      string(spec.TypeConfig.AutomationLevel),
		SkillSlugs:           skillSlugs,
		SourcePodKey:         &pod.PodKey,
		WorkerSpecSnapshotID: &snapshotID,
	}
	return s.Create(ctx, createReq)
}

func (s *Service) resolveSnapshotSkillSlugs(
	ctx context.Context,
	organizationID int64,
	skillIDs []int64,
) ([]string, error) {
	if len(skillIDs) == 0 {
		return nil, nil
	}
	if s.skills == nil {
		return nil, ErrWorkerSpecSkillUnavailable
	}
	slugs := make([]string, 0, len(skillIDs))
	for _, skillID := range skillIDs {
		row, err := s.skills.GetAnyByID(ctx, skillID)
		if err != nil {
			if errors.Is(err, skilldomain.ErrNotFound) {
				return nil, errors.Join(
					ErrWorkerSpecSnapshotMismatch,
					fmt.Errorf("skill %d is missing", skillID),
				)
			}
			return nil, errors.Join(ErrWorkerSpecSkillUnavailable, err)
		}
		if row == nil || !row.VisibleTo(organizationID) || row.Slug == "" {
			return nil, errors.Join(
				ErrWorkerSpecSnapshotMismatch,
				fmt.Errorf("skill %d is not visible to organization", skillID),
			)
		}
		slugs = append(slugs, row.Slug)
	}
	return slugs, nil
}
