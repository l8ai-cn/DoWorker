package orchestrationcontrol

import (
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
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

type ModelBindingReferenceFilter struct {
	WorkerType       slugkit.Slug
	ProtocolAdapters []string
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
	if filter.EnvironmentBundle != nil {
		if err := filter.validateEnvironmentBundle(); err != nil {
			return err
		}
	}
	if filter.ModelBinding != nil {
		if err := filter.validateModelBinding(); err != nil {
			return err
		}
	}
	return nil
}

func (filter ResourceListFilter) validateEnvironmentBundle() error {
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

func (filter ResourceListFilter) validateModelBinding() error {
	if filter.Kind != resource.KindModelBinding {
		return fmt.Errorf(
			"%w: model binding filter requires ModelBinding kind",
			control.ErrInvalid,
		)
	}
	if err := slugkit.Validate(filter.ModelBinding.WorkerType.String()); err != nil {
		return fmt.Errorf("%w: invalid model binding worker type", control.ErrInvalid)
	}
	seenAdapters := make(map[string]struct{}, len(filter.ModelBinding.ProtocolAdapters))
	for _, adapter := range filter.ModelBinding.ProtocolAdapters {
		if err := slugkit.Validate(adapter); err != nil {
			return fmt.Errorf("%w: invalid model protocol adapter", control.ErrInvalid)
		}
		if _, exists := seenAdapters[adapter]; exists {
			return fmt.Errorf("%w: duplicate model protocol adapter", control.ErrInvalid)
		}
		seenAdapters[adapter] = struct{}{}
	}
	return nil
}

func invalidEnvironmentBundleFilter() error {
	return fmt.Errorf(
		"%w: invalid environment bundle reference context",
		control.ErrInvalid,
	)
}
