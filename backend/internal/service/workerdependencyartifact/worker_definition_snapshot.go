package workerdependencyartifact

import (
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/merge"
	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/anthropics/agentsmesh/agentfile/serialize"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
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
