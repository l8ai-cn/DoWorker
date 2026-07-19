package workerdependencyartifact

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func validateWorkspace(
	spec workerspec.Spec,
	document workerdependency.Document,
) error {
	workspace := spec.Workspace
	if err := validateRepository(workspace, document.Repository); err != nil {
		return err
	}
	if !sameIDSet(workspace.SkillIDs, skillIDs(document.Skills)) {
		return fmt.Errorf("worker dependency Skills do not match WorkerSpec")
	}
	if !sameKnowledgeMounts(workspace.KnowledgeMounts, document.KnowledgeBases) {
		return fmt.Errorf("worker dependency KnowledgeBases do not match WorkerSpec")
	}
	if !sameRuntimeBundles(workspace.EnvBundleIDs, document.RuntimeBundles) {
		return fmt.Errorf("worker dependency runtime bundles do not match WorkerSpec")
	}
	if !sameConfigDocuments(
		workspace.ConfigDocumentBindings,
		document.RuntimeBundles,
	) {
		return fmt.Errorf("worker dependency config documents do not match WorkerSpec")
	}
	if !sameSecretReferences(spec.TypeConfig.SecretRefs, document.SecretReferences) {
		return fmt.Errorf("worker dependency secret references do not match WorkerSpec")
	}
	return nil
}

func validateRepository(
	workspace workerspec.Workspace,
	repository *workerdependency.Repository,
) error {
	if workspace.RepositoryID == nil {
		if repository != nil {
			return fmt.Errorf("worker dependency repository is absent from WorkerSpec")
		}
		return nil
	}
	if repository == nil ||
		repository.Pin.DomainID != *workspace.RepositoryID ||
		repository.Branch != workspace.Branch {
		return fmt.Errorf("worker dependency repository does not match WorkerSpec")
	}
	return nil
}

func skillIDs(skills []workerdependency.Skill) []int64 {
	ids := make([]int64, len(skills))
	for index := range skills {
		ids[index] = skills[index].Pin.DomainID
	}
	return ids
}

func sameIDSet(left, right []int64) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[int64]struct{}, len(left))
	for _, id := range left {
		if _, exists := seen[id]; exists {
			return false
		}
		seen[id] = struct{}{}
	}
	actual := make(map[int64]struct{}, len(right))
	for _, id := range right {
		if _, exists := seen[id]; !exists {
			return false
		}
		if _, exists := actual[id]; exists {
			return false
		}
		actual[id] = struct{}{}
	}
	return true
}

func sameKnowledgeMounts(
	mounts []workerspec.KnowledgeMount,
	dependencies []workerdependency.KnowledgeBase,
) bool {
	if len(mounts) != len(dependencies) {
		return false
	}
	expected := make(map[int64]workerspec.KnowledgeMountMode, len(mounts))
	for _, mount := range mounts {
		if _, exists := expected[mount.KnowledgeBaseID]; exists {
			return false
		}
		expected[mount.KnowledgeBaseID] = mount.Mode
	}
	actual := make(map[int64]struct{}, len(dependencies))
	for _, dependency := range dependencies {
		if _, exists := actual[dependency.Pin.DomainID]; exists {
			return false
		}
		actual[dependency.Pin.DomainID] = struct{}{}
		if expected[dependency.Pin.DomainID] != dependency.Mode {
			return false
		}
	}
	return true
}
