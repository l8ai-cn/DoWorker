package expert

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

func (s *Service) refreshExpertWorkerSpec(
	ctx context.Context,
	before, after *expertdom.Expert,
) (int64, error) {
	if before.WorkerSpecSnapshotID == nil {
		return 0, nil
	}
	if unsupportedSnapshotFieldsChanged(before, after) {
		return 0, ErrExpertSnapshotUpdateUnsupported
	}
	if !editableSnapshotFieldsChanged(before, after) {
		return 0, nil
	}
	if s.workerSpecs == nil || s.workerSpecWriter == nil {
		return 0, ErrExpertSnapshotUnavailable
	}
	snapshot, err := s.workerSpecs.GetByID(
		ctx,
		after.OrganizationID,
		*before.WorkerSpecSnapshotID,
	)
	if err != nil {
		return 0, errors.Join(ErrExpertSnapshotUnavailable, err)
	}
	spec, err := updatedExpertSpec(snapshot.Spec, after)
	if err != nil {
		return 0, err
	}
	resolved, err := specservice.NewResolvedSnapshot(after.OrganizationID, spec)
	if err != nil {
		return 0, err
	}
	created, err := s.workerSpecWriter.Create(ctx, resolved)
	if err != nil {
		return 0, err
	}
	after.WorkerSpecSnapshotID = &created.ID
	return created.ID, nil
}

func updatedExpertSpec(
	spec specdomain.Spec,
	expert *expertdom.Expert,
) (specdomain.Spec, error) {
	values := map[string]any{}
	if len(expert.ConfigOverrides) > 0 {
		if err := json.Unmarshal(expert.ConfigOverrides, &values); err != nil {
			return specdomain.Spec{}, err
		}
	}
	spec.Workspace.Instructions = optionalStringValue(expert.Prompt)
	spec.TypeConfig.InteractionMode = specdomain.InteractionMode(
		expert.InteractionMode,
	)
	spec.TypeConfig.AutomationLevel = specdomain.AutomationLevel(
		expert.AutomationLevel,
	)
	spec.TypeConfig.Values = values
	return specdomain.NormalizeAndValidate(spec)
}

func unsupportedSnapshotFieldsChanged(
	before, after *expertdom.Expert,
) bool {
	return before.AgentSlug != after.AgentSlug ||
		!reflect.DeepEqual(before.RunnerID, after.RunnerID) ||
		!reflect.DeepEqual(before.RepositoryID, after.RepositoryID) ||
		!reflect.DeepEqual(before.BranchName, after.BranchName) ||
		before.Perpetual != after.Perpetual ||
		!reflect.DeepEqual(before.UsedEnvBundles, after.UsedEnvBundles) ||
		!reflect.DeepEqual(before.SkillSlugs, after.SkillSlugs) ||
		!reflect.DeepEqual(before.KnowledgeMounts, after.KnowledgeMounts) ||
		!reflect.DeepEqual(before.AgentfileLayer, after.AgentfileLayer)
}

func editableSnapshotFieldsChanged(
	before, after *expertdom.Expert,
) bool {
	return optionalStringValue(before.Prompt) != optionalStringValue(after.Prompt) ||
		before.InteractionMode != after.InteractionMode ||
		before.AutomationLevel != after.AutomationLevel ||
		!reflect.DeepEqual(before.ConfigOverrides, after.ConfigOverrides)
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func (s *Service) cleanupExpertSnapshot(
	ctx context.Context,
	organizationID, snapshotID int64,
) {
	if snapshotID == 0 || s.workerSpecWriter == nil {
		return
	}
	cleanupCtx, cancel := marketCleanupContext(ctx)
	defer cancel()
	if err := s.workerSpecWriter.Delete(
		cleanupCtx,
		organizationID,
		snapshotID,
	); err != nil {
		s.logger.Warn(
			"expert: compensating workerspec snapshot delete failed",
			"snapshot_id", snapshotID,
			"error", err,
		)
	}
}
