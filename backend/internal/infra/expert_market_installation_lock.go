package infra

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"gorm.io/gorm"
)

var ErrExpertMarketLockRequiresPostgres = errors.New(
	"expert market lock requires PostgreSQL",
)

type ExpertMarketInstallationLocker struct {
	db *gorm.DB
}

func NewExpertMarketInstallationLocker(
	db *gorm.DB,
) *ExpertMarketInstallationLocker {
	return &ExpertMarketInstallationLocker{db: db}
}

func (locker *ExpertMarketInstallationLocker) WithinMarketInstallationLock(
	ctx context.Context,
	organizationID, applicationID int64,
	apply func() error,
) error {
	if locker == nil || locker.db == nil || apply == nil {
		return expertmarket.ErrConflict
	}
	if locker.db.Dialector.Name() != "postgres" {
		return fmt.Errorf(
			"%w: %s",
			ErrExpertMarketLockRequiresPostgres,
			locker.db.Dialector.Name(),
		)
	}
	key := marketInstallationLockKey(organizationID, applicationID)
	return locker.withinLock(ctx, key, apply)
}

func (locker *ExpertMarketInstallationLocker) WithinMarketApplicationLock(
	ctx context.Context,
	applicationID int64,
	apply func() error,
) error {
	if applicationID <= 0 {
		return expertmarket.ErrConflict
	}
	return locker.withinLock(
		ctx,
		marketApplicationLockKey(applicationID),
		apply,
	)
}

func (locker *ExpertMarketInstallationLocker) withinLock(
	ctx context.Context,
	key int64,
	apply func() error,
) error {
	if locker == nil || locker.db == nil || apply == nil {
		return expertmarket.ErrConflict
	}
	if locker.db.Dialector.Name() != "postgres" {
		return fmt.Errorf(
			"%w: %s",
			ErrExpertMarketLockRequiresPostgres,
			locker.db.Dialector.Name(),
		)
	}
	return locker.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"SELECT pg_advisory_xact_lock(?)",
			key,
		).Error; err != nil {
			return err
		}
		return apply()
	})
}

func marketInstallationLockKey(
	organizationID, applicationID int64,
) int64 {
	hasher := fnv.New64a()
	var buffer [16]byte
	binary.LittleEndian.PutUint64(buffer[:8], uint64(organizationID))
	binary.LittleEndian.PutUint64(buffer[8:], uint64(applicationID))
	_, _ = hasher.Write(buffer[:])
	return int64(hasher.Sum64())
}

func marketApplicationLockKey(applicationID int64) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte("market-application"))
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[:], uint64(applicationID))
	_, _ = hasher.Write(buffer[:])
	return int64(hasher.Sum64())
}
