package airesource

import domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"

func connectionView(connection *domain.Connection, canManage bool, resources []*domain.ModelResource) ConnectionView {
	resourceViews := make([]ResourceView, len(resources))
	for index, resource := range resources {
		resourceViews[index] = resourceView(resource)
	}
	return ConnectionView{
		ID: connection.ID, OwnerScope: connection.OwnerScope, OwnerID: connection.OwnerID,
		Identifier: connection.Identifier, ProviderKey: connection.ProviderKey, Name: connection.Name,
		BaseURL: connection.BaseURL, ConfiguredFields: append([]string(nil), connection.ConfiguredFields...),
		Status: connection.Status, IsEnabled: connection.IsEnabled, LastValidatedAt: connection.LastValidatedAt,
		ValidationError: connection.ValidationError, CanManage: canManage, Resources: resourceViews,
	}
}

func resourceView(resource *domain.ModelResource) ResourceView {
	return ResourceView{
		ID: resource.ID, ProviderConnectionID: resource.ProviderConnectionID, Identifier: resource.Identifier,
		ModelID: resource.ModelID, DisplayName: resource.DisplayName,
		Modalities: append([]domain.Modality(nil), resource.Modalities...), Capabilities: append([]domain.Capability(nil), resource.Capabilities...),
		DefaultModalities: append([]domain.Modality(nil), resource.DefaultModalities...), Status: resource.Status,
		IsEnabled: resource.IsEnabled, LastValidatedAt: resource.LastValidatedAt, ValidationError: resource.ValidationError,
		UsageSummary: resource.UsageSummary,
	}
}
