package workerspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func Normalize(spec Spec) (Spec, error) {
	normalized := spec
	normalized.Runtime.WorkerType.DefinitionHash = strings.TrimSpace(
		spec.Runtime.WorkerType.DefinitionHash,
	)
	normalized.Runtime.Image, normalized.Placement = NormalizeRuntimePlacement(
		spec.Runtime.Image,
		spec.Placement,
	)

	values, err := cloneJSONValues(spec.TypeConfig.Values)
	if err != nil {
		return Spec{}, fmt.Errorf("type_config.values: %w", err)
	}
	normalized.TypeConfig.Values = values
	normalized.TypeConfig.SecretRefs = cloneSecretReferences(spec.TypeConfig.SecretRefs)
	if normalized.TypeConfig.InteractionMode == "" {
		normalized.TypeConfig.InteractionMode = InteractionModePTY
	}
	if normalized.TypeConfig.AutomationLevel == "" {
		normalized.TypeConfig.AutomationLevel = AutomationLevelAutonomous
	}

	normalized.Workspace.RepositoryID = cloneInt64Pointer(spec.Workspace.RepositoryID)
	normalized.Workspace.Branch = strings.TrimSpace(spec.Workspace.Branch)
	normalized.Workspace.SkillIDs = append([]int64{}, spec.Workspace.SkillIDs...)
	sort.Slice(normalized.Workspace.SkillIDs, func(i, j int) bool {
		return normalized.Workspace.SkillIDs[i] < normalized.Workspace.SkillIDs[j]
	})
	normalized.Workspace.KnowledgeMounts = append(
		[]KnowledgeMount{},
		spec.Workspace.KnowledgeMounts...,
	)
	for index := range normalized.Workspace.KnowledgeMounts {
		if normalized.Workspace.KnowledgeMounts[index].Mode == "" {
			normalized.Workspace.KnowledgeMounts[index].Mode = KnowledgeMountReadOnly
		}
	}
	sort.Slice(normalized.Workspace.KnowledgeMounts, func(i, j int) bool {
		return normalized.Workspace.KnowledgeMounts[i].KnowledgeBaseID <
			normalized.Workspace.KnowledgeMounts[j].KnowledgeBaseID
	})
	normalized.Workspace.EnvBundleIDs = append(
		[]RuntimeEnvBundleID{},
		spec.Workspace.EnvBundleIDs...,
	)
	normalized.Workspace.Instructions = strings.TrimSpace(spec.Workspace.Instructions)
	normalized.Workspace.InitialTask = strings.TrimSpace(spec.Workspace.InitialTask)

	if normalized.Lifecycle.TerminationPolicy == "" {
		normalized.Lifecycle.TerminationPolicy = TerminationPolicyManual
	}
	normalized.Metadata.Alias = strings.TrimSpace(spec.Metadata.Alias)
	normalized.Metadata.SourceExpertID = cloneInt64Pointer(spec.Metadata.SourceExpertID)
	return normalized, nil
}

func cloneJSONValues(values map[string]any) (map[string]any, error) {
	if values == nil {
		return map[string]any{}, nil
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var cloned map[string]any
	if err := decoder.Decode(&cloned); err != nil {
		return nil, err
	}
	return cloned, nil
}

func cloneSecretReferences(
	references map[string]SecretReference,
) map[string]SecretReference {
	cloned := make(map[string]SecretReference, len(references))
	for field, reference := range references {
		cloned[field] = reference
	}
	return cloned
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func clonePlacement(placement Placement) Placement {
	cloned := placement
	resources := placement.ResourceProfile.Resources
	cloned.ResourceProfile.Resources.GPURequest = cloneUint32Pointer(resources.GPURequest)
	cloned.ResourceProfile.Resources.GPULimit = cloneUint32Pointer(resources.GPULimit)
	return cloned
}

func cloneUint32Pointer(value *uint32) *uint32 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
