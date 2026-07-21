package agentpod

import (
	"sort"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func workerSpecResourceRequirements(
	spec *specdomain.Spec,
) ([]int64, []int64, []specdomain.ConfigDocumentBinding) {
	if spec == nil {
		return nil, nil, nil
	}
	envBundleSet := make(
		map[int64]struct{},
		len(spec.Workspace.EnvBundleIDs)+len(spec.TypeConfig.SecretRefs),
	)
	for _, id := range spec.Workspace.EnvBundleIDs {
		envBundleSet[int64(id)] = struct{}{}
	}
	for _, reference := range spec.TypeConfig.SecretRefs {
		envBundleSet[reference.ID] = struct{}{}
	}
	envBundleIDs := make([]int64, 0, len(envBundleSet))
	for id := range envBundleSet {
		envBundleIDs = append(envBundleIDs, id)
	}
	sort.Slice(envBundleIDs, func(left, right int) bool {
		return envBundleIDs[left] < envBundleIDs[right]
	})
	skillIDs := append([]int64{}, spec.Workspace.SkillIDs...)
	sort.Slice(skillIDs, func(left, right int) bool {
		return skillIDs[left] < skillIDs[right]
	})
	configBindings := append(
		[]specdomain.ConfigDocumentBinding{},
		spec.Workspace.ConfigDocumentBindings...,
	)
	return envBundleIDs, skillIDs, configBindings
}

func workerSpecSecretEnvBundleIDs(spec *specdomain.Spec) []int64 {
	if spec == nil {
		return nil
	}
	envBundleSet := make(map[int64]struct{}, len(spec.TypeConfig.SecretRefs))
	for _, reference := range spec.TypeConfig.SecretRefs {
		envBundleSet[reference.ID] = struct{}{}
	}
	envBundleIDs := make([]int64, 0, len(envBundleSet))
	for id := range envBundleSet {
		envBundleIDs = append(envBundleIDs, id)
	}
	sort.Slice(envBundleIDs, func(left, right int) bool {
		return envBundleIDs[left] < envBundleIDs[right]
	})
	return envBundleIDs
}
