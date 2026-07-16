package skill

import (
	"context"
	"errors"
	"fmt"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
)

type packagePersist func(
	store skilldom.Repository,
	pkg *extensionsvc.PackagedSkill,
) (bool, error)

type packageCompensate func(
	store skilldom.Repository,
	pkg *extensionsvc.PackagedSkill,
	cause error,
) error

func (s *Service) publishPreparedPackage(
	ctx context.Context,
	store skilldom.Repository,
	prepared *extensionsvc.PreparedSkill,
	persist packagePersist,
	compensate packageCompensate,
) (conflict bool, err error) {
	err = store.WithPackageLock(ctx, prepared.StorageKey, func(locked skilldom.Repository) error {
		pkg, storeErr := s.packager.StorePrepared(ctx, prepared)
		if storeErr != nil {
			return fmt.Errorf("skill: store package: %w", storeErr)
		}
		var persistErr error
		conflict, persistErr = persist(locked, pkg)
		if persistErr != nil {
			return compensate(locked, pkg, persistErr)
		}
		if conflict {
			return compensate(locked, pkg, nil)
		}
		return nil
	})
	return conflict, err
}

func (s *Service) cleanupCreatedPackage(
	ctx context.Context,
	store skilldom.Repository,
	pkg *extensionsvc.PackagedSkill,
	cause error,
) error {
	if pkg == nil || pkg.StorageKey == "" || !pkg.Created {
		return cause
	}
	cleanupCtx := context.WithoutCancel(ctx)
	referenced, err := store.IsPackageReferenced(cleanupCtx, pkg.StorageKey)
	if err != nil {
		return errors.Join(cause, fmt.Errorf("skill: check package references: %w", err))
	}
	if referenced {
		return cause
	}
	if err := s.packager.DeletePackage(cleanupCtx, pkg.StorageKey); err != nil {
		return errors.Join(
			cause,
			fmt.Errorf("skill: delete unreferenced package: %w", err),
		)
	}
	return cause
}
