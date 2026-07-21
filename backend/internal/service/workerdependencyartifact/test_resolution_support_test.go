package workerdependencyartifact

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/google/uuid"
)

func workerSpecModelBinding(model workerdependency.Model) workerspec.ModelBinding {
	return workerspec.ModelBinding{
		ResourceID: model.Pin.DomainID, ResourceRevision: model.ResourceRevision,
		ConnectionID:       model.ConnectionID,
		ConnectionRevision: model.ConnectionRevision,
		ProviderKey:        model.ProviderKey, ProtocolAdapter: model.ProtocolAdapter,
		ModelID: model.ModelID,
	}
}

func modelResolution(
	reference control.ResolvedReference,
	model workerdependency.Model,
) ModelResolution {
	return ModelResolution{
		ResourceResolution: resourceResolution(reference, model.Pin.DomainID),
		ResourceRevision:   model.ResourceRevision,
		ConnectionID:       model.ConnectionID,
		ConnectionRevision: model.ConnectionRevision,
		ProviderKey:        model.ProviderKey, ProtocolAdapter: model.ProtocolAdapter,
		ModelID: model.ModelID, BaseURL: model.BaseURL,
		Modalities:   append([]airesource.Modality{}, model.Modalities...),
		Capabilities: append([]airesource.Capability{}, model.Capabilities...),
	}
}

func resourceResolution(
	reference control.ResolvedReference,
	domainID int64,
) ResourceResolution {
	return ResourceResolution{reference: reference, domainID: domainID}
}

func runtimeValueResolutions(
	values []workerdependency.RuntimeValue,
) []RuntimeValueResolution {
	result := make([]RuntimeValueResolution, len(values))
	for index, value := range values {
		result[index] = RuntimeValueResolution{Name: value.Name, Value: value.Value}
	}
	return result
}

func resolvedReference(kind, name string) control.ResolvedReference {
	identity := kind + "\x00" + name
	return control.ResolvedReference{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		},
		Namespace: slugkit.MustNewForTest("team-alpha"),
		Name:      slugkit.MustNewForTest(name),
		UID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte(identity)).String(),
		Revision:  1,
		Digest:    workerdependency.TextDigest(identity),
	}
}

func dependencyPin(
	reference control.ResolvedReference,
	domainID int64,
) workerdependency.ResourcePin {
	return workerdependency.ResourcePin{
		Reference: referenceFromResolved(reference),
		DomainID:  domainID,
	}
}

func referenceFromResolved(
	reference control.ResolvedReference,
) resource.Reference {
	return resource.Reference{
		APIVersion: reference.APIVersion,
		Kind:       reference.Kind,
		Namespace:  reference.Namespace,
		Name:       reference.Name,
		UID:        reference.UID,
		Revision:   reference.Revision,
		Digest:     reference.Digest,
	}
}

func definitionSnapshot(
	t testHelper,
	comprehensive bool,
) workerdefinition.Definition {
	t.Helper()
	source := `{"schema_version":1,"slug":"codex-cli","definition_version":"1","executable":"codex","adapter_id":"codex-app-server","interaction_modes":["pty","acp"],"model_requirement":{"required":false,"protocol_adapters":[]},"tool_model_requirements":[{"id":"video-generation","provider_keys":["doubao"],"protocol_adapters":["openai-compatible"],"modality":"video","capability":"video-generation","environment":{"api_key":"VIDEO_API_KEY","base_url":"VIDEO_BASE_URL","model_id":"VIDEO_MODEL_ID"}}],"credential_bindings":[],"config_documents":[],"image":{"runtime":"codex-cli","version_probe":["codex","--version"]}}`
	agentfile := "AGENT codex\nENV VIDEO_API_KEY SECRET OPTIONAL\n" +
		"ENV VIDEO_BASE_URL TEXT OPTIONAL\nENV VIDEO_MODEL_ID TEXT OPTIONAL\nMODE pty\n"
	if comprehensive {
		source = `{"schema_version":1,"slug":"codex-cli","definition_version":"1","executable":"codex","adapter_id":"codex-app-server","interaction_modes":["pty","acp"],"model_requirement":{"required":true,"protocol_adapters":["openai-compatible"]},"tool_model_requirements":[{"id":"video-generation","provider_keys":["doubao"],"protocol_adapters":["openai-compatible"],"modality":"video","capability":"video-generation","environment":{"api_key":"VIDEO_API_KEY","base_url":"VIDEO_BASE_URL","model_id":"VIDEO_MODEL_ID"}}],"credential_bindings":[{"id":"cursor","source":{"kind":"credential_bundle","ref":"cursor"},"target":{"kind":"env","name":"CURSOR_API_KEY"}}],"config_documents":[{"id":"settings","format":"json","target_path":".codex/settings.json","required":true}],"image":{"runtime":"codex-cli","version_probe":["codex","--version"]}}`
		agentfile = "AGENT codex\nENV VIDEO_API_KEY SECRET OPTIONAL\n" +
			"ENV VIDEO_BASE_URL TEXT OPTIONAL\nENV VIDEO_MODEL_ID TEXT OPTIONAL\n" +
			"ENV CURSOR_API_KEY SECRET OPTIONAL\nMODE pty\n"
	}
	definition, err := workerdefinition.ParseSnapshot([]byte(source), agentfile)
	if err != nil {
		t.Fatalf("parse test worker definition: %v", err)
	}
	return definition
}
