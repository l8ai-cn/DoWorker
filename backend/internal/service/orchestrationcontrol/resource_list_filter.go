package orchestrationcontrol

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type EnvironmentBundlePurpose string

const (
	EnvironmentBundlePurposeRuntime    EnvironmentBundlePurpose = "runtime"
	EnvironmentBundlePurposeConfig     EnvironmentBundlePurpose = "config"
	EnvironmentBundlePurposeCredential EnvironmentBundlePurpose = "credential"
)

type EnvironmentBundleReferenceFilter struct {
	Purpose            EnvironmentBundlePurpose
	WorkerType         slugkit.Slug
	TargetName         string
	ModelManagedFields []string
}

func (filter ResourceListFilter) Validate(scope control.Scope) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if filter.Limit <= 0 || filter.Limit > 100 || filter.Offset < 0 {
		return fmt.Errorf("%w: invalid resource list pagination", control.ErrInvalid)
	}
	if filter.Kind != "" {
		if err := (resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       filter.Kind,
		}).Validate(); err != nil {
			return fmt.Errorf("%w: invalid resource list kind", control.ErrInvalid)
		}
	}
	if filter.EnvironmentBundle == nil {
		return nil
	}
	if filter.Kind != resource.KindEnvironmentBundle {
		return fmt.Errorf(
			"%w: environment bundle filter requires EnvironmentBundle kind",
			control.ErrInvalid,
		)
	}
	switch filter.EnvironmentBundle.Purpose {
	case EnvironmentBundlePurposeRuntime:
		if filter.EnvironmentBundle.TargetName != "" {
			return invalidEnvironmentBundleFilter()
		}
	case EnvironmentBundlePurposeConfig:
		if filter.EnvironmentBundle.TargetName != "" ||
			len(filter.EnvironmentBundle.ModelManagedFields) != 0 {
			return invalidEnvironmentBundleFilter()
		}
	case EnvironmentBundlePurposeCredential:
		if err := workerspec.ValidateConfigField(
			filter.EnvironmentBundle.TargetName,
		); err != nil || len(filter.EnvironmentBundle.ModelManagedFields) != 0 {
			return invalidEnvironmentBundleFilter()
		}
	default:
		return fmt.Errorf(
			"%w: invalid environment bundle purpose",
			control.ErrInvalid,
		)
	}
	seenFields := make(map[string]struct{}, len(
		filter.EnvironmentBundle.ModelManagedFields,
	))
	for _, field := range filter.EnvironmentBundle.ModelManagedFields {
		if err := workerspec.ValidateConfigField(field); err != nil {
			return invalidEnvironmentBundleFilter()
		}
		if _, exists := seenFields[field]; exists {
			return invalidEnvironmentBundleFilter()
		}
		seenFields[field] = struct{}{}
	}
	if err := slugkit.Validate(filter.EnvironmentBundle.WorkerType.String()); err != nil {
		return fmt.Errorf(
			"%w: invalid environment bundle worker type",
			control.ErrInvalid,
		)
	}
	return nil
}

func invalidEnvironmentBundleFilter() error {
	return fmt.Errorf(
		"%w: invalid environment bundle reference context",
		control.ErrInvalid,
	)
}
