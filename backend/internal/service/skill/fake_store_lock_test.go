package skill

import (
	"context"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

func (f *fakeStore) WithMutationLock(
	_ context.Context,
	_ int64,
	mutate func(skilldom.Repository) error,
) error {
	return mutate(f)
}
