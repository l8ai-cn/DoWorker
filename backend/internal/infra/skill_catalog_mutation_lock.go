package infra

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"gorm.io/gorm"
)

func (r *SkillCatalogRepository) WithMutationLock(
	ctx context.Context,
	id int64,
	mutate func(skilldom.Repository) error,
) error {
	if r.db.Name() != "postgres" {
		return mutate(r)
	}
	key := skillMutationLockKey(id)
	return r.db.WithContext(ctx).Connection(func(lockedDB *gorm.DB) error {
		if err := lockedDB.Exec("SELECT pg_advisory_lock(?)", key).Error; err != nil {
			return fmt.Errorf("skill: acquire mutation lock: %w", err)
		}
		mutationErr := mutate(&SkillCatalogRepository{db: lockedDB})
		unlockDB := lockedDB.Session(&gorm.Session{Context: context.WithoutCancel(ctx)})
		unlockErr := unlockDB.Exec("SELECT pg_advisory_unlock(?)", key).Error
		if unlockErr != nil {
			return errors.Join(
				mutationErr,
				fmt.Errorf("skill: release mutation lock: %w", unlockErr),
			)
		}
		return mutationErr
	})
}

func skillMutationLockKey(id int64) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte("skill-mutation"))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(strconv.FormatInt(id, 10)))
	return int64(hash.Sum64())
}
