package airesourceconnect

import (
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	service "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	aiv1 "github.com/anthropics/agentsmesh/proto/gen/go/ai_resource/v1"
)

func providerToProto(provider domain.ProviderDefinition) *aiv1.ProviderDefinition {
	fields := make([]*aiv1.CredentialField, len(provider.CredentialFields))
	for index, field := range provider.CredentialFields {
		fields[index] = &aiv1.CredentialField{Key: field.Key, Label: field.Label, Secret: field.Secret, Required: field.Required}
	}
	return &aiv1.ProviderDefinition{
		Key: provider.Key.String(), DisplayName: provider.DisplayName, Modalities: modalityStrings(provider.Modalities),
		CredentialFields: fields, DefaultBaseUrl: provider.DefaultBaseURL, ProtocolAdapter: provider.ProtocolAdapter,
		SupportsCustomEndpoint: provider.SupportsCustomEndpoint, SupportsModelDiscovery: provider.SupportsModelDiscovery,
	}
}

func connectionToProto(connection service.ConnectionView) *aiv1.ProviderConnection {
	resources := make([]*aiv1.ModelResource, len(connection.Resources))
	for index := range connection.Resources {
		resources[index] = resourceToProto(connection.Resources[index])
	}
	return &aiv1.ProviderConnection{
		Id: connection.ID, OwnerScope: string(connection.OwnerScope), Identifier: connection.Identifier.String(),
		ProviderKey: connection.ProviderKey.String(), Name: connection.Name, BaseUrl: connection.BaseURL,
		ConfiguredFields: append([]string(nil), connection.ConfiguredFields...), Status: string(connection.Status),
		IsEnabled: connection.IsEnabled, LastValidatedAt: timeString(connection.LastValidatedAt),
		ValidationError: connection.ValidationError, CanManage: connection.CanManage, Resources: resources,
	}
}

func resourceToProto(resource service.ResourceView) *aiv1.ModelResource {
	return &aiv1.ModelResource{
		Id: resource.ID, ProviderConnectionId: resource.ProviderConnectionID, Identifier: resource.Identifier.String(),
		ModelId: resource.ModelID, DisplayName: resource.DisplayName, Modalities: modalityStrings(resource.Modalities),
		Capabilities: capabilityStrings(resource.Capabilities), DefaultModalities: modalityStrings(resource.DefaultModalities),
		Status: string(resource.Status), IsEnabled: resource.IsEnabled, LastValidatedAt: timeString(resource.LastValidatedAt),
		ValidationError: resource.ValidationError, UsageSummary: usageToProto(resource.UsageSummary),
	}
}

func usageToProto(usage *domain.UsageSummary) *aiv1.UsageSummary {
	if usage == nil {
		return nil
	}
	return &aiv1.UsageSummary{
		QuotaTotal: usage.QuotaTotal, UsageTotal: usage.UsageTotal, Remaining: usage.Remaining,
		Unit: usage.Unit, Period: usage.Period, MeasuredAt: timeString(usage.MeasuredAt),
	}
}

func effectiveToProto(items []service.EffectiveResourceView) []*aiv1.EffectiveResource {
	result := make([]*aiv1.EffectiveResource, len(items))
	for index, item := range items {
		result[index] = &aiv1.EffectiveResource{Connection: connectionToProto(item.Connection), Resource: resourceToProto(item.Resource), Selectable: item.Selectable, BlockingReason: string(item.BlockingReason)}
	}
	return result
}

func connectionsToProto(items []service.ConnectionView) []*aiv1.ProviderConnection {
	result := make([]*aiv1.ProviderConnection, len(items))
	for index := range items {
		result[index] = connectionToProto(items[index])
	}
	return result
}

func modalitiesFromProto(values []string) []domain.Modality {
	result := make([]domain.Modality, len(values))
	for index, value := range values {
		result[index] = domain.Modality(value)
	}
	return result
}

func capabilitiesFromProto(values []string) []domain.Capability {
	result := make([]domain.Capability, len(values))
	for index, value := range values {
		result[index] = domain.Capability(value)
	}
	return result
}

func modalityStrings(values []domain.Modality) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func capabilityStrings(values []domain.Capability) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func timeString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339Nano)
	return &formatted
}
