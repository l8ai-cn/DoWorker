package skill

import (
	"context"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
)

func (f *fakeStore) WithMutationLock(
	_ context.Context,
	_ int64,
	mutate func(skilldom.Repository) error,
) error {
	return mutate(f)
}

func (f *fakeStore) WithPackageLock(
	_ context.Context,
	_ string,
	mutate func(skilldom.Repository) error,
) error {
	return mutate(f)
}

func (f *fakeStore) IsPackageReferenced(_ context.Context, storageKey string) (bool, error) {
	for _, row := range f.rows {
		if row.StorageKey == storageKey {
			return true, nil
		}
	}
	return false, nil
}
