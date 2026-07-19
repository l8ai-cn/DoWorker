package workerdependencyartifact

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/merge"
	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/anthropics/agentsmesh/agentfile/serialize"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent/automation"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func buildWorkerSnapshot(
	definition workerdefinition.Definition,
	layer string,
	spec workerspec.Spec,
	specDigest string,
) (workerdependency.Worker, error) {
	if err := workerdefinition.ValidateIntegrity(definition); err != nil {
		return workerdependency.Worker{}, err
	}
	if definition.Slug != spec.Runtime.WorkerType.Slug.String() ||
		definition.DefinitionHash != spec.Runtime.WorkerType.DefinitionHash {
		return workerdependency.Worker{}, fmt.Errorf(
			"worker definition does not match Plan WorkerSpec",
		)
	}
	workerType, err := slugkit.NewFromTrusted(definition.Slug)
	if err != nil {
		return workerdependency.Worker{}, fmt.Errorf("worker definition slug: %w", err)
	}
	adapterID, err := slugkit.NewFromTrusted(definition.AdapterID)
	if err != nil {
		return workerdependency.Worker{}, fmt.Errorf("worker definition adapter id: %w", err)
	}
	layer, err = workerAgentfileLayer(definition.Slug, layer, spec)
	if err != nil {
		return workerdependency.Worker{}, err
	}
	source, err := mergedAgentfile(definition.AgentFile, layer)
	if err != nil {
		return workerdependency.Worker{}, err
	}
	policy := workerdefinition.BuildEnvironmentBundlePolicy(definition)
	return workerdependency.Worker{
		WorkerType: workerType, AdapterID: adapterID,
		SpecVersion: spec.Version, SpecDigest: specDigest,
		DefinitionHash: definition.DefinitionHash,
		ModelManagedFields: append(
			[]string{},
			policy.ModelManagedFields...,
		),
		CredentialBundleFields: append(
			[]string{},
			policy.CredentialBundleFields...,
		),
		AgentfileSource: source, AgentfileSourceDigest: workerdependency.TextDigest(source),
	}, nil
}

func workerAgentfileLayer(
	workerSlug, layer string, spec workerspec.Spec,
) (string, error) {
	program, parseErrors := parser.Parse(layer)
	if len(parseErrors) != 0 {
		return "", fmt.Errorf("worker AgentFile layer is invalid: %s", parseErrors[0])
	}
	explicit := make(map[string]struct{})
	for _, declaration := range program.Declarations {
		if config, ok := declaration.(*parser.ConfigDecl); ok {
			explicit[config.Name] = struct{}{}
		}
	}
	output := automation.AdapterFor(workerSlug).Apply(
		string(spec.TypeConfig.AutomationLevel),
	)
	keys := make([]string, 0, len(output.ConfigOverrides))
	for key := range output.ConfigOverrides {
		if _, found := explicit[key]; found {
			continue
		}
		if _, found := spec.TypeConfig.Values[key]; found {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		layer = appendLayerConfig(layer, key, output.ConfigOverrides[key])
	}
	return layer, nil
}

func appendLayerConfig(layer, key, value string) string {
	line := fmt.Sprintf("CONFIG %s = %q", key, value)
	if strings.TrimSpace(layer) == "" {
		return line
	}
	return strings.TrimRight(layer, "\n") + "\n" + line
}

func mergedAgentfile(base, layer string) (string, error) {
	baseProgram, parseErrors := parser.Parse(base)
	if len(parseErrors) != 0 {
		return "", fmt.Errorf("worker definition AgentFile is invalid: %s", parseErrors[0])
	}
	if strings.TrimSpace(layer) == "" {
		return serialize.Serialize(baseProgram), nil
	}
	layerProgram, parseErrors := parser.Parse(layer)
	if len(parseErrors) != 0 {
		return "", fmt.Errorf("worker AgentFile layer is invalid: %s", parseErrors[0])
	}
	for _, declaration := range layerProgram.Declarations {
		if _, isEnvironment := declaration.(*parser.EnvDecl); isEnvironment {
			return "", fmt.Errorf("worker AgentFile layer must not declare ENV fields")
		}
	}
	merge.Merge(baseProgram, layerProgram)
	return serialize.Serialize(baseProgram), nil
}
