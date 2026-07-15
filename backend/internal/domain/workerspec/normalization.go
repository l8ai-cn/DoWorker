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
	normalized.Runtime.ModelBinding.ModelID = strings.TrimSpace(
		spec.Runtime.ModelBinding.ModelID,
	)
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

	normalized.Workspace.RepositoryID = cloneInt64Pointer(spec.Workspace.RepositoryID)
	normalized.Workspace.Branch = strings.TrimSpace(spec.Workspace.Branch)
	normalized.Workspace.SkillIDs = append([]int64{}, spec.Workspace.SkillIDs...)
	sort.Slice(normalized.Workspace.SkillIDs, func(i, j int) bool {
		return normalized.Workspace.SkillIDs[i] < normalized.Workspace.SkillIDs[j]
	})
	normalized.Workspace.SkillPackages = append(
		[]SkillPackageBinding{},
		spec.Workspace.SkillPackages...,
	)
	for index := range normalized.Workspace.SkillPackages {
		pkg := &normalized.Workspace.SkillPackages[index]
		pkg.Slug = strings.TrimSpace(pkg.Slug)
		pkg.ContentSHA = strings.TrimSpace(pkg.ContentSHA)
		pkg.StorageKey = strings.TrimSpace(pkg.StorageKey)
	}
	sort.Slice(normalized.Workspace.SkillPackages, func(i, j int) bool {
		return normalized.Workspace.SkillPackages[i].SkillID <
			normalized.Workspace.SkillPackages[j].SkillID
	})
	normalized.Workspace.KnowledgeMounts = append(
		[]KnowledgeMount{},
		spec.Workspace.KnowledgeMounts...,
	)
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

	normalized.Metadata.Alias = strings.TrimSpace(spec.Metadata.Alias)
	normalized.Metadata.SourceExpertID = cloneInt64Pointer(spec.Metadata.SourceExpertID)
	return normalized, nil
}

func cloneJSONValues(values map[string]any) (map[string]any, error) {
	if values == nil {
		return nil, nil
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
	if references == nil {
		return nil
	}
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
