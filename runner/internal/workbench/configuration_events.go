package workbench

import (
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func (m *Mapper) SessionInitialized(
	configuration acp.Configuration,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	mutations := []*agentworkbenchv2.RunnerWorkbenchMutation{
		capabilitiesMutation(configuration),
	}
	if current := sessionConfiguration(
		configuration.Model,
		configuration.PermissionMode,
	); current != nil {
		mutations = append(mutations, configurationMutation(current))
	}
	return m.batchLocked(configuration, mutations...)
}

func (m *Mapper) ConfigurationChanged(
	update acp.ConfigUpdate,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	current := sessionConfiguration(update.Model, update.PermissionMode)
	if current == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.batchLocked(update, configurationMutation(current))
}

func capabilitiesMutation(
	configuration acp.Configuration,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	commandSchemas := []*agentworkbenchv2.CapabilityDescriptor{
		commandCapability("send_prompt", "session.send"),
		commandCapability("interrupt", "session.interrupt"),
		commandCapability("resolve_permission", "session.permission.resolve"),
	}
	if len(configuration.SupportedModels) > 0 ||
		len(configuration.SupportedPermissionModes) > 0 {
		commandSchemas = append(
			commandSchemas,
			commandCapability("change_configuration", "session.configure"),
		)
	}
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Capabilities{
			Capabilities: &agentworkbenchv2.SupportCapabilities{
				ProtocolVersion: "2",
				CommandSchemas:  commandSchemas,
				Models:          append([]string(nil), configuration.SupportedModels...),
				PermissionModes: append([]string(nil), configuration.SupportedPermissionModes...),
				ArtifactOperations: append(
					[]string{"artifact.download"},
					withoutArtifactDownload(
						configuration.SupportedArtifactActions,
					)...,
				),
				History: true,
			},
		},
	}
}

func withoutArtifactDownload(actions []string) []string {
	filtered := make([]string, 0, len(actions))
	for _, action := range actions {
		if action != "" && action != "artifact.download" {
			filtered = append(filtered, action)
		}
	}
	return filtered
}

func commandCapability(
	semanticKey string,
	action string,
) *agentworkbenchv2.CapabilityDescriptor {
	return &agentworkbenchv2.CapabilityDescriptor{
		Namespace: "proto.agent_workbench.v2", SemanticKey: semanticKey,
		SchemaVersion: "2", Actions: []string{action},
	}
}

func configurationMutation(
	configuration *agentworkbenchv2.SessionConfiguration,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Configuration{
			Configuration: configuration,
		},
	}
}

func sessionConfiguration(
	model string,
	permissionMode string,
) *agentworkbenchv2.SessionConfiguration {
	if model == "" && permissionMode == "" {
		return nil
	}
	return &agentworkbenchv2.SessionConfiguration{
		Model:          optionalConfigurationValue(model),
		PermissionMode: optionalConfigurationValue(permissionMode),
	}
}

func optionalConfigurationValue(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
