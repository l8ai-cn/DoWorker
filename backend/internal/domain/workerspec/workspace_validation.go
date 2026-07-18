package workerspec

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

const maxBranchRunes = 255

func validateWorkspace(workspace Workspace) error {
	switch {
	case workspace.RepositoryID == nil && workspace.Branch != "":
		return fmt.Errorf("workspace repository is required when branch is set")
	case workspace.RepositoryID != nil && *workspace.RepositoryID <= 0:
		return fmt.Errorf("workspace repository id must be positive")
	case workspace.RepositoryID != nil && workspace.Branch == "":
		return fmt.Errorf("workspace branch is required with a repository")
	case strings.TrimSpace(workspace.Branch) != workspace.Branch:
		return fmt.Errorf("workspace branch must be normalized")
	case utf8.RuneCountInString(workspace.Branch) > maxBranchRunes:
		return fmt.Errorf("workspace branch exceeds %d characters", maxBranchRunes)
	}
	if err := validateUniqueIDs("workspace skill_ids", workspace.SkillIDs); err != nil {
		return err
	}
	if err := validateSkillPackages(workspace.SkillIDs, workspace.SkillPackages); err != nil {
		return err
	}
	if err := validateKnowledgeMounts(workspace.KnowledgeMounts); err != nil {
		return err
	}
	ids := make([]int64, len(workspace.EnvBundleIDs))
	for index, id := range workspace.EnvBundleIDs {
		ids[index] = int64(id)
	}
	if err := validateUniqueIDs("workspace env_bundle_ids", ids); err != nil {
		return err
	}
	return validateConfigDocumentBindings(workspace.ConfigDocumentBindings)
}

func validateConfigDocumentBindings(bindings []ConfigDocumentBinding) error {
	documents := make(map[string]struct{}, len(bindings))
	bundles := make(map[int64]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding.DocumentID == "" ||
			strings.TrimSpace(binding.DocumentID) != binding.DocumentID {
			return fmt.Errorf("workspace config document id must be normalized")
		}
		if _, exists := documents[binding.DocumentID]; exists {
			return fmt.Errorf(
				"workspace config_document_bindings contains duplicate document %q",
				binding.DocumentID,
			)
		}
		if binding.ConfigBundleID <= 0 {
			return fmt.Errorf(
				"workspace config document %q bundle id must be positive",
				binding.DocumentID,
			)
		}
		if _, exists := bundles[binding.ConfigBundleID]; exists {
			return fmt.Errorf(
				"workspace config_document_bindings contains duplicate bundle id %d",
				binding.ConfigBundleID,
			)
		}
		documents[binding.DocumentID] = struct{}{}
		bundles[binding.ConfigBundleID] = struct{}{}
	}
	return nil
}

func validateKnowledgeMounts(mounts []KnowledgeMount) error {
	seen := make(map[int64]struct{}, len(mounts))
	for _, mount := range mounts {
		if mount.KnowledgeBaseID <= 0 {
			return fmt.Errorf("workspace knowledge_mounts id must be positive")
		}
		if _, exists := seen[mount.KnowledgeBaseID]; exists {
			return fmt.Errorf(
				"workspace knowledge_mounts contains duplicate id %d",
				mount.KnowledgeBaseID,
			)
		}
		seen[mount.KnowledgeBaseID] = struct{}{}
		switch mount.Mode {
		case KnowledgeMountReadOnly, KnowledgeMountReadWrite:
		default:
			return fmt.Errorf(
				"workspace knowledge_mounts id %d has invalid mode %q",
				mount.KnowledgeBaseID,
				mount.Mode,
			)
		}
	}
	return nil
}

func validateUniqueIDs(field string, ids []int64) error {
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("%s values must be positive", field)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("%s contains duplicate id %d", field, id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateLifecycle(lifecycle Lifecycle) error {
	switch lifecycle.TerminationPolicy {
	case TerminationPolicyOnIdle:
		if lifecycle.IdleTimeoutMinutes == 0 {
			return fmt.Errorf("lifecycle idle timeout must be positive for idle policy")
		}
	case TerminationPolicyManual, TerminationPolicyOnCompleted:
		if lifecycle.IdleTimeoutMinutes != 0 {
			return fmt.Errorf(
				"lifecycle idle timeout must be zero for %q policy",
				lifecycle.TerminationPolicy,
			)
		}
	default:
		return fmt.Errorf(
			"invalid lifecycle termination policy %q",
			lifecycle.TerminationPolicy,
		)
	}
	return nil
}
