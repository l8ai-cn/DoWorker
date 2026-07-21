package podconnect

import (
	"context"
	"errors"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	podv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/pod/v1"
)

var errWorkerSpecLoaderUnavailable = errors.New("worker spec snapshot loader unavailable")

func (s *Server) applyWorkerSkills(
	ctx context.Context,
	organizationID int64,
	items []*podv1.Pod,
) error {
	requested := make(map[int64]struct{})
	for _, item := range items {
		snapshotID := item.GetWorkerSpecSnapshotId()
		if snapshotID != 0 {
			requested[snapshotID] = struct{}{}
		}
	}
	if len(requested) == 0 {
		return nil
	}
	if s.workerSpecs == nil {
		return errWorkerSpecLoaderUnavailable
	}

	skillsBySnapshot := make(map[int64][]string, len(requested))
	if len(requested) == 1 {
		for snapshotID := range requested {
			snapshot, err := s.workerSpecs.GetByID(ctx, organizationID, snapshotID)
			if err != nil {
				return err
			}
			skillsBySnapshot[snapshotID] = workerSkillSlugs(snapshot)
		}
	} else {
		ids := make([]int64, 0, len(requested))
		for snapshotID := range requested {
			ids = append(ids, snapshotID)
		}
		snapshots, err := s.workerSpecs.GetByIDs(ctx, organizationID, ids)
		if err != nil {
			return err
		}
		for _, snapshot := range snapshots {
			if _, ok := requested[snapshot.ID]; ok {
				skillsBySnapshot[snapshot.ID] = workerSkillSlugs(snapshot)
			}
		}
	}

	for _, item := range items {
		snapshotID := item.GetWorkerSpecSnapshotId()
		if snapshotID == 0 {
			continue
		}
		skills, ok := skillsBySnapshot[snapshotID]
		if !ok {
			return workerspec.ErrNotFound
		}
		item.WorkerSkillSlugs = skills
	}
	return nil
}

func workerSkillSlugs(snapshot workerspec.Snapshot) []string {
	skills := make([]string, 0, len(snapshot.Spec.Workspace.SkillPackages))
	for _, binding := range snapshot.Spec.Workspace.SkillPackages {
		skills = append(skills, binding.Slug)
	}
	return skills
}
