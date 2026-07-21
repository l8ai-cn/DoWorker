package infra

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func filterModelBindingReferences(
	query *gorm.DB,
	scope control.Scope,
	filter service.ResourceListFilter,
) (*gorm.DB, error) {
	if filter.ModelBinding == nil {
		return query, nil
	}
	dialect := query.Name()
	entityID, err := modelBindingEntityIDExpression(dialect)
	if err != nil {
		return nil, err
	}
	providerKeys := providerKeysForAdapters(
		filter.ModelBinding.ProtocolAdapters,
	)
	if len(providerKeys) == 0 {
		return nil, control.ErrInvalid
	}
	filtered := query.
		Joins(`
JOIN orchestration_resource_revisions AS active_model_binding_revision
  ON active_model_binding_revision.organization_id = orchestration_resources.organization_id
 AND active_model_binding_revision.resource_id = orchestration_resources.id
 AND active_model_binding_revision.revision = orchestration_resources.active_revision`).
		Joins(fmt.Sprintf(`
JOIN model_resources AS referenced_model_resource
  ON referenced_model_resource.id = %s`, entityID)).
		Joins(`
JOIN provider_connections AS referenced_model_connection
  ON referenced_model_connection.id = referenced_model_resource.provider_connection_id`).
		Where(
			"referenced_model_connection.provider_key IN ?",
			providerKeys,
		).
		Where(
			"referenced_model_resource.is_enabled = ? AND "+
				"referenced_model_connection.is_enabled = ?",
			true,
			true,
		).
		Where(
			"referenced_model_resource.status = ? AND "+
				"referenced_model_connection.status = ?",
			airesource.ConnectionStatusValid,
			airesource.ConnectionStatusValid,
		).
		Where(
			`(
  (referenced_model_connection.owner_scope = ? AND referenced_model_connection.owner_id = ?)
  OR
  (referenced_model_connection.owner_scope = ? AND referenced_model_connection.owner_id = ?)
)`,
			airesource.OwnerScopeUser,
			scope.ActorID,
			airesource.OwnerScopeOrg,
			scope.OrganizationID,
		)
	return filterWorkerPrimaryModelCapabilities(
		filtered,
		dialect,
	)
}

func modelBindingEntityIDExpression(dialect string) (string, error) {
	switch dialect {
	case "postgres":
		return "CAST(active_model_binding_revision.canonical_spec ->> " +
			"'resourceId' AS BIGINT)", nil
	case "sqlite":
		return "CAST(json_extract(active_model_binding_revision.canonical_spec, " +
			"'$.resourceId') AS INTEGER)", nil
	default:
		return "", fmt.Errorf(
			"%w: model binding reference filtering does not support %s",
			service.ErrUnavailable,
			dialect,
		)
	}
}

func filterWorkerPrimaryModelCapabilities(
	query *gorm.DB,
	dialect string,
) (*gorm.DB, error) {
	switch dialect {
	case "postgres":
		return query.
			Where(
				"referenced_model_resource.modalities @> ?::jsonb",
				`["chat"]`,
			).
			Where(
				"referenced_model_resource.capabilities @> ?::jsonb",
				`["text-generation"]`,
			), nil
	case "sqlite":
		return query.
			Where(
				"EXISTS (SELECT 1 FROM json_each(referenced_model_resource.modalities) WHERE value = ?)",
				"chat",
			).
			Where(
				"EXISTS (SELECT 1 FROM json_each(referenced_model_resource.capabilities) WHERE value = ?)",
				"text-generation",
			), nil
	default:
		return nil, fmt.Errorf(
			"%w: model binding reference filtering does not support %s",
			service.ErrUnavailable,
			dialect,
		)
	}
}

func providerKeysForAdapters(adapters []string) []string {
	allowed := make(map[string]struct{}, len(adapters))
	for _, adapter := range adapters {
		allowed[adapter] = struct{}{}
	}
	keys := make([]string, 0, len(allowed))
	for _, provider := range airesource.Providers() {
		if _, exists := allowed[provider.ProtocolAdapter]; exists {
			keys = append(keys, provider.Key.String())
		}
	}
	return keys
}
