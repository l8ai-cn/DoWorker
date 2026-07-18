package infra

import (
	"fmt"

	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func filterEnvironmentBundleData(
	query *gorm.DB,
	dialect string,
	filter service.EnvironmentBundleReferenceFilter,
) (*gorm.DB, error) {
	switch filter.Purpose {
	case service.EnvironmentBundlePurposeRuntime:
		for _, field := range filter.ModelManagedFields {
			var err error
			query, err = excludeEnvironmentBundleDataKey(query, dialect, field)
			if err != nil {
				return nil, err
			}
		}
		return query, nil
	case service.EnvironmentBundlePurposeCredential:
		return requireEnvironmentBundleDataKey(query, dialect, filter.TargetName)
	case service.EnvironmentBundlePurposeConfig:
		return query, nil
	default:
		return nil, fmt.Errorf(
			"%w: unsupported environment bundle purpose",
			service.ErrUnavailable,
		)
	}
}

func excludeEnvironmentBundleDataKey(
	query *gorm.DB,
	dialect string,
	field string,
) (*gorm.DB, error) {
	switch dialect {
	case "postgres":
		return query.Where(
			"NOT jsonb_exists(referenced_environment_bundle.data, ?)",
			field,
		), nil
	case "sqlite":
		return query.Where(`
NOT EXISTS (
  SELECT 1
  FROM json_each(referenced_environment_bundle.data) AS excluded_environment_field
  WHERE excluded_environment_field.key = ?
)`, field), nil
	default:
		return nil, unsupportedEnvironmentBundleDataDialect(dialect)
	}
}

func requireEnvironmentBundleDataKey(
	query *gorm.DB,
	dialect string,
	field string,
) (*gorm.DB, error) {
	switch dialect {
	case "postgres":
		return query.Where(
			"jsonb_exists(referenced_environment_bundle.data, ?)",
			field,
		), nil
	case "sqlite":
		return query.Where(`
EXISTS (
  SELECT 1
  FROM json_each(referenced_environment_bundle.data) AS required_environment_field
  WHERE required_environment_field.key = ?
)`, field), nil
	default:
		return nil, unsupportedEnvironmentBundleDataDialect(dialect)
	}
}

func unsupportedEnvironmentBundleDataDialect(dialect string) error {
	return fmt.Errorf(
		"%w: environment bundle data filtering does not support %s",
		service.ErrUnavailable,
		dialect,
	)
}
