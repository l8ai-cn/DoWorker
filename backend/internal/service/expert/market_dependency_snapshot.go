package expert

import (
	"encoding/json"
	"errors"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func validateMarketSkillDependencies(
	raw json.RawMessage,
	packages []specdomain.SkillPackageBinding,
) error {
	var expected []MarketSkillDependency
	if err := decodeStrictJSON(raw, &expected); err != nil || expected == nil {
		return errors.Join(ErrMarketSnapshotInvalid, err)
	}
	if len(expected) != len(packages) {
		return ErrMarketSnapshotInvalid
	}
	expectedByID := make(map[int64]MarketSkillDependency, len(expected))
	for _, dependency := range expected {
		if dependency.SkillID <= 0 || dependency.Version <= 0 ||
			dependency.ContentSHA == "" || dependency.StorageKey == "" ||
			dependency.PackageSize < 0 {
			return ErrMarketSnapshotInvalid
		}
		if err := slugkit.Validate(dependency.Slug); err != nil {
			return errors.Join(ErrMarketSnapshotInvalid, err)
		}
		if _, exists := expectedByID[dependency.SkillID]; exists {
			return ErrMarketSnapshotInvalid
		}
		expectedByID[dependency.SkillID] = dependency
	}
	for _, pkg := range packages {
		dependency, exists := expectedByID[pkg.SkillID]
		if !exists || pkg.Slug != dependency.Slug ||
			pkg.Version != dependency.Version ||
			pkg.ContentSHA != dependency.ContentSHA ||
			pkg.StorageKey != dependency.StorageKey ||
			pkg.PackageSize != dependency.PackageSize {
			return ErrMarketSnapshotInvalid
		}
	}
	return nil
}
