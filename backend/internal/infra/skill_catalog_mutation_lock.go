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
		return executeSkillMutationLock(
			func() error {
				if err := lockedDB.Exec("SELECT pg_advisory_lock(?)", key).Error; err != nil {
					return fmt.Errorf("skill: acquire mutation lock: %w", err)
				}
				return nil
			},
			func() error {
				unlockDB := lockedDB.Session(&gorm.Session{
					Context: context.WithoutCancel(ctx),
				})
				if err := unlockDB.Exec("SELECT pg_advisory_unlock(?)", key).Error; err != nil {
					return fmt.Errorf("skill: release mutation lock: %w", err)
				}
				return nil
			},
			func() error {
				return mutate(&SkillCatalogRepository{db: lockedDB})
			},
		)
	})
}

func executeSkillMutationLock(
	acquire func() error,
	release func() error,
	mutate func() error,
) (err error) {
	if err := acquire(); err != nil {
		return err
	}
	defer func() {
		mutationPanic := recover()
		releaseErr, releasePanic := callSkillMutationUnlock(release)
		if mutationPanic != nil {
			panic(mutationPanic)
		}
		if releasePanic != nil {
			panic(releasePanic)
		}
		err = errors.Join(err, releaseErr)
	}()
	return mutate()
}

func callSkillMutationUnlock(release func() error) (err error, panicValue any) {
	defer func() { panicValue = recover() }()
	return release(), nil
}

func skillMutationLockKey(id int64) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte("skill-mutation"))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(strconv.FormatInt(id, 10)))
	return int64(hash.Sum64())
}
