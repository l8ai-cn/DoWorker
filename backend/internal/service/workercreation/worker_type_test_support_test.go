package workercreation

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	agentdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
)

type workerTypeAgentProvider struct {
	agent *agentdomain.Agent
	err   error
}

func (provider *workerTypeAgentProvider) GetAgent(
	context.Context,
	string,
) (*agentdomain.Agent, error) {
	return provider.agent, provider.err
}

func (provider *workerTypeAgentProvider) ListBuiltinAgents(
	context.Context,
) ([]*agentdomain.Agent, error) {
	if provider.err != nil {
		return nil, provider.err
	}
	if provider.agent == nil {
		return nil, nil
	}
	return []*agentdomain.Agent{provider.agent}, nil
}

func activeWorkerTypeAgent(source string) *agentdomain.Agent {
	return activeWorkerTypeAgentFor("codex-cli", "codex", source)
}

func activeWorkerTypeAgentFor(slug, executable, source string) *agentdomain.Agent {
	return &agentdomain.Agent{
		Slug: slug, Name: slug, Executable: executable, AdapterID: "test-adapter",
		AgentfileSource: &source, IsActive: true, SupportedModes: "pty,acp",
	}
}

func workerDefinition(
	slug, executable, source string,
	modes ...string,
) workerdefinition.Definition {
	bindings := make([]map[string]any, 0)
	for _, field := range agentfileCredentialFields(source) {
		if field == "OPENAI_API_KEY" {
			bindings = append(bindings, credentialBindingDocument(
				"openai", "model_resource", "openai-compatible", field,
			))
			continue
		}
		bindings = append(bindings, credentialBindingDocument(
			strings.ToLower(strings.ReplaceAll(field, "_", "-")),
			"credential_bundle", slug, field,
		))
	}
	return parseWorkerDefinition(
		definitionDocument(
			slug, executable, "test-adapter", modes,
			modelRequirementDocument(), bindings, []any{},
		),
		source,
	)
}

func noModelWorkerDefinition(
	slug, executable, source string,
	modes ...string,
) workerdefinition.Definition {
	definition := workerDefinition(slug, executable, source, modes...)
	return parseWorkerDefinition(
		definitionDocument(
			slug, executable, definition.AdapterID, modes,
			map[string]any{"required": false, "protocol_adapters": []string{}},
			credentialBundleBindingDocuments(slug, source), []any{},
		),
		source,
	)
}

func toolModelWorkerDefinition(
	slug, executable, source string,
	toolModels []map[string]any,
	modes ...string,
) workerdefinition.Definition {
	items := make([]any, len(toolModels))
	for index, item := range toolModels {
		items[index] = item
	}
	return parseWorkerDefinition(
		definitionDocument(
			slug, executable, "test-adapter", modes, modelRequirementDocument(),
			credentialBundleBindingDocumentsExcept(
				slug, source, toolModelEnvironmentFields(toolModels),
			),
			items,
		),
		source,
	)
}

func parseWorkerDefinition(
	document map[string]any,
	source string,
) workerdefinition.Definition {
	raw, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	definition, err := workerdefinition.ParseSnapshot(raw, source)
	if err != nil {
		panic(err)
	}
	return definition
}

func definitionDocument(
	slug, executable, adapterID string,
	modes []string,
	modelRequirement map[string]any,
	bindings []map[string]any,
	toolModels []any,
) map[string]any {
	return map[string]any{
		"schema_version":                1,
		"slug":                          slug,
		"definition_version":            "1",
		"executable":                    executable,
		"adapter_id":                    adapterID,
		"interaction_modes":             modes,
		"model_requirement":             modelRequirement,
		"tool_model_requirements":       toolModels,
		"credential_bindings":           bindings,
		"credential_requirement_groups": []any{},
		"config_documents":              []any{},
		"image": map[string]any{
			"runtime": slug, "version_probe": []string{executable, "--version"},
		},
	}
}

func modelRequirementDocument() map[string]any {
	return map[string]any{
		"required": true, "protocol_adapters": []string{"openai-compatible"},
	}
}

func agentfileCredentialFields(source string) []string {
	matches := regexp.MustCompile(`(?m)^ENV\s+([A-Z][A-Z0-9_]*)\s+SECRET\b`).
		FindAllStringSubmatch(source, -1)
	fields := make([]string, 0, len(matches))
	for _, match := range matches {
		fields = append(fields, match[1])
	}
	return fields
}

func credentialBundleBindingDocuments(slug, source string) []map[string]any {
	return credentialBundleBindingDocumentsExcept(slug, source, map[string]struct{}{})
}

func credentialBundleBindingDocumentsExcept(
	slug, source string,
	excluded map[string]struct{},
) []map[string]any {
	bindings := make([]map[string]any, 0)
	for _, field := range agentfileCredentialFields(source) {
		if _, skip := excluded[field]; skip {
			continue
		}
		bindings = append(bindings, credentialBindingDocument(
			strings.ToLower(strings.ReplaceAll(field, "_", "-")),
			"credential_bundle", slug, field,
		))
	}
	return bindings
}

func toolModelEnvironmentFields(items []map[string]any) map[string]struct{} {
	fields := map[string]struct{}{}
	for _, item := range items {
		raw, ok := item["environment"].(map[string]any)
		if !ok {
			continue
		}
		for _, key := range []string{"api_key", "base_url", "model_id"} {
			if value, ok := raw[key].(string); ok {
				fields[value] = struct{}{}
			}
		}
	}
	return fields
}

func credentialBindingDocument(id, kind, ref, target string) map[string]any {
	return map[string]any{
		"id":     id,
		"source": map[string]any{"kind": kind, "ref": ref},
		"target": map[string]any{"kind": "env", "name": target},
	}
}

func credentialBinding(
	id, ref, target string,
) workerdefinition.CredentialBinding {
	return workerdefinition.CredentialBinding{
		ID: id,
		Source: workerdefinition.CredentialSource{
			Kind: "credential_bundle", Ref: ref,
		},
		Target: workerdefinition.CredentialTarget{Kind: "env", Name: target},
	}
}
