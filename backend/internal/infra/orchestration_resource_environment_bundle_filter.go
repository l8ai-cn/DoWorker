package infra

import (
	"fmt"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func filterEnvironmentBundleReferences(
	query *gorm.DB,
	scope control.Scope,
	filter service.ResourceListFilter,
) (*gorm.DB, error) {
	if filter.EnvironmentBundle == nil {
		return query, nil
	}
	dialect := query.Name()
	entityID, err := environmentBundleEntityIDExpression(dialect)
	if err != nil {
		return nil, err
	}
	kinds, err := environmentBundlePurposeKinds(
		filter.EnvironmentBundle.Purpose,
	)
	if err != nil {
		return nil, err
	}
	if err := validateEnvironmentBundleReferenceIntegrity(
		query,
		dialect,
		entityID,
	); err != nil {
		return nil, err
	}
	filtered := query.
		Joins(`
JOIN orchestration_resource_revisions AS active_environment_revision
  ON active_environment_revision.organization_id = orchestration_resources.organization_id
 AND active_environment_revision.resource_id = orchestration_resources.id
 AND active_environment_revision.revision = orchestration_resources.active_revision`).
		Joins(fmt.Sprintf(`
JOIN env_bundles AS referenced_environment_bundle
  ON referenced_environment_bundle.id = %s`, entityID)).
		Where("referenced_environment_bundle.is_active = ?", true).
		Where("referenced_environment_bundle.kind IN ?", kinds).
		Where(
			`(
  (referenced_environment_bundle.owner_scope = ? AND referenced_environment_bundle.owner_id = ?)
  OR
  (referenced_environment_bundle.owner_scope = ? AND referenced_environment_bundle.owner_id = ?)
)`,
			envbundle.OwnerScopeUser,
			scope.ActorID,
			envbundle.OwnerScopeOrg,
			scope.OrganizationID,
		).
		Where(
			"(referenced_environment_bundle.agent_slug IS NULL OR "+
				"referenced_environment_bundle.agent_slug = ?)",
			filter.EnvironmentBundle.WorkerType.String(),
		)
	return filterEnvironmentBundleData(
		filtered,
		dialect,
		*filter.EnvironmentBundle,
	)
}

func validateEnvironmentBundleReferenceIntegrity(
	query *gorm.DB,
	dialect string,
	entityID string,
) error {
	validID, err := environmentBundleIDValidityPredicate(dialect)
	if err != nil {
		return err
	}
	var invalid int64
	err = query.Session(&gorm.Session{}).
		Joins(`
LEFT JOIN orchestration_resource_revisions AS integrity_environment_revision
  ON integrity_environment_revision.organization_id = orchestration_resources.organization_id
 AND integrity_environment_revision.resource_id = orchestration_resources.id
 AND integrity_environment_revision.revision = orchestration_resources.active_revision`).
		Joins(fmt.Sprintf(`
LEFT JOIN env_bundles AS integrity_environment_bundle
  ON integrity_environment_bundle.id = %s`,
			environmentBundleEntityIDForAlias(entityID, "integrity_environment_revision"),
		)).
		Where(
			"integrity_environment_revision.id IS NULL OR NOT (" + validID + ") " +
				"OR integrity_environment_bundle.id IS NULL",
		).
		Count(&invalid).Error
	if err != nil {
		return fmt.Errorf(
			"%w: validate environment bundle resource bindings: %v",
			service.ErrUnavailable,
			err,
		)
	}
	if invalid > 0 {
		return fmt.Errorf(
			"%w: environment bundle resources contain invalid active bindings",
			service.ErrUnavailable,
		)
	}
	return nil
}

func environmentBundleEntityIDExpression(dialect string) (string, error) {
	switch dialect {
	case "postgres":
		return "CAST(active_environment_revision.canonical_spec ->> " +
			"'environmentBundleId' AS BIGINT)", nil
	case "sqlite":
		return "CAST(json_extract(active_environment_revision.canonical_spec, " +
			"'$.environmentBundleId') AS INTEGER)", nil
	default:
		return "", fmt.Errorf(
			"%w: environment bundle reference filtering does not support %s",
			service.ErrUnavailable,
			dialect,
		)
	}
}

func environmentBundleEntityIDForAlias(expression string, alias string) string {
	return strings.ReplaceAll(
		expression,
		"active_environment_revision",
		alias,
	)
}

func environmentBundleIDValidityPredicate(dialect string) (string, error) {
	switch dialect {
	case "postgres":
		return "jsonb_typeof(integrity_environment_revision.canonical_spec -> " +
			"'environmentBundleId') = 'number' AND " +
			"(integrity_environment_revision.canonical_spec ->> " +
			"'environmentBundleId') ~ '^[1-9][0-9]*$'", nil
	case "sqlite":
		return "json_type(integrity_environment_revision.canonical_spec, " +
			"'$.environmentBundleId') = 'integer' AND " +
			"json_extract(integrity_environment_revision.canonical_spec, " +
			"'$.environmentBundleId') > 0", nil
	default:
		return "", fmt.Errorf(
			"%w: environment bundle reference filtering does not support %s",
			service.ErrUnavailable,
			dialect,
		)
	}
}

func environmentBundlePurposeKinds(
	purpose service.EnvironmentBundlePurpose,
) ([]string, error) {
	switch purpose {
	case service.EnvironmentBundlePurposeRuntime:
		return []string{envbundle.KindRuntime, envbundle.KindShared}, nil
	case service.EnvironmentBundlePurposeConfig:
		return []string{envbundle.KindConfig}, nil
	case service.EnvironmentBundlePurposeCredential:
		return []string{envbundle.KindCredential}, nil
	default:
		return nil, control.ErrInvalid
	}
}
