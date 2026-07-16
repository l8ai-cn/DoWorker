package expert

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func (s *Service) validateMarketReleaseRuntimeContract(
	ctx context.Context,
	application *expertmarket.Application,
	release *expertmarket.Release,
) error {
	latest, err := s.market.GetLatestReleaseByApplication(ctx, application.ID)
	if errors.Is(err, expertmarket.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	_, previousWorker, err := decodeMarketReleaseSnapshots(latest)
	if err != nil {
		return err
	}
	_, nextWorker, err := decodeMarketReleaseSnapshots(release)
	if err != nil {
		return err
	}
	if sameMarketRuntimeContract(previousWorker.Spec, nextWorker.Spec) {
		return nil
	}
	return errors.Join(
		ErrMarketSnapshotInvalid,
		fmt.Errorf("market releases cannot change the worker runtime contract"),
	)
}

func sameMarketRuntimeContract(previous, next specdomain.Spec) bool {
	return previous.Runtime.WorkerType == next.Runtime.WorkerType &&
		equalMarketToolModelRoles(
			previous.Runtime.ToolModelBindings,
			next.Runtime.ToolModelBindings,
		)
}

func equalMarketToolModelRoles(
	previous, next []specdomain.ToolModelBinding,
) bool {
	if len(previous) != len(next) {
		return false
	}
	previousRoles := marketToolModelRoles(previous)
	nextRoles := marketToolModelRoles(next)
	for index := range previousRoles {
		if previousRoles[index] != nextRoles[index] {
			return false
		}
	}
	return true
}

func marketToolModelRoles(bindings []specdomain.ToolModelBinding) []string {
	roles := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		roles = append(roles, binding.Role.String())
	}
	sort.Strings(roles)
	return roles
}
