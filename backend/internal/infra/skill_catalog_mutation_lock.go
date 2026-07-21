package infra

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	"gorm.io/gorm"
)

type SkillMutationPanic struct {
	Value     any
	UnlockErr error
}

func (p *SkillMutationPanic) Error() string {
	return fmt.Sprintf("skill mutation panic: %v; unlock failed: %v", p.Value, p.UnlockErr)
}

func (p *SkillMutationPanic) Unwrap() []error {
	errs := make([]error, 0, 2)
	if mutationErr, ok := p.Value.(error); ok {
		errs = append(errs, mutationErr)
	}
	if p.UnlockErr != nil {
		errs = append(errs, p.UnlockErr)
	}
	return errs
}

func (r *SkillCatalogRepository) WithMutationLock(
	ctx context.Context,
	id int64,
	mutate func(skilldom.Repository) error,
) error {
	return r.withAdvisoryLock(ctx, skillMutationLockKey(id), mutate)
}

func (r *SkillCatalogRepository) WithPackageLock(
	ctx context.Context,
	storageKey string,
	mutate func(skilldom.Repository) error,
) error {
	return r.withAdvisoryLock(
		ctx,
		skillAdvisoryLockKey("skill-package", storageKey),
		mutate,
	)
}

func (r *SkillCatalogRepository) withAdvisoryLock(
	ctx context.Context,
	key int64,
	mutate func(skilldom.Repository) error,
) error {
	if r.db.Name() != "postgres" {
		return mutate(r)
	}
	if r.sessionBound {
		return r.executeAdvisoryLock(ctx, key, mutate)
	}
	return r.db.WithContext(ctx).Connection(func(lockedDB *gorm.DB) error {
		bound := &SkillCatalogRepository{db: lockedDB, sessionBound: true}
		return bound.executeAdvisoryLock(ctx, key, mutate)
	})
}

func (r *SkillCatalogRepository) executeAdvisoryLock(
	ctx context.Context,
	key int64,
	mutate func(skilldom.Repository) error,
) error {
	lockedDB := r.db.WithContext(ctx)
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
		func() error { return mutate(r) },
	)
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
		releasePanic, releaseErr := callSkillMutationUnlock(release)
		unlockErr := skillMutationUnlockError(releaseErr, releasePanic)
		if mutationPanic != nil {
			if unlockErr != nil {
				panic(&SkillMutationPanic{
					Value:     mutationPanic,
					UnlockErr: unlockErr,
				})
			}
			panic(mutationPanic)
		}
		if releasePanic != nil {
			if err != nil {
				panic(&SkillMutationPanic{
					Value:     err,
					UnlockErr: unlockErr,
				})
			}
			panic(releasePanic)
		}
		err = errors.Join(err, releaseErr)
	}()
	return mutate()
}

func callSkillMutationUnlock(release func() error) (panicValue any, err error) {
	defer func() { panicValue = recover() }()
	return nil, release()
}

func skillMutationUnlockError(releaseErr error, releasePanic any) error {
	if releasePanic == nil {
		return releaseErr
	}
	if panicErr, ok := releasePanic.(error); ok {
		return errors.Join(releaseErr, panicErr)
	}
	return errors.Join(
		releaseErr,
		fmt.Errorf("skill mutation unlock panic: %v", releasePanic),
	)
}

func skillMutationLockKey(id int64) int64 {
	return skillAdvisoryLockKey("skill-mutation", strconv.FormatInt(id, 10))
}

func skillAdvisoryLockKey(namespace, value string) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(namespace))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(value))
	return int64(hash.Sum64())
}
