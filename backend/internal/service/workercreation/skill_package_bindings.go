package workercreation

import (
	"fmt"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func skillPackageIndex(
	bindings []specdomain.SkillPackageBinding,
) map[int64]specdomain.SkillPackageBinding {
	index := make(map[int64]specdomain.SkillPackageBinding, len(bindings))
	for _, binding := range bindings {
		index[binding.SkillID] = binding
	}
	return index
}

func requiredSkillPackage(
	index map[int64]specdomain.SkillPackageBinding,
	id int64,
) (specdomain.SkillPackageBinding, error) {
	binding, ok := index[id]
	if !ok || binding.SkillID != id {
		return specdomain.SkillPackageBinding{}, fmt.Errorf(
			"WorkerTemplate artifact skill package %d is missing",
			id,
		)
	}
	return binding, nil
}
