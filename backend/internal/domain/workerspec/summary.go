package workerspec

import (
	"fmt"
	"unicode/utf8"
)

type Summary struct {
	Version             Version      `json:"version"`
	ModelBinding        ModelBinding `json:"model_binding"`
	WorkerType          WorkerType   `json:"worker_type"`
	RuntimeImage        RuntimeImage `json:"runtime_image"`
	Placement           Placement    `json:"placement"`
	Alias               string       `json:"alias"`
	RepositoryID        *int64       `json:"repository_id,omitempty"`
	Branch              string       `json:"branch"`
	SkillCount          uint32       `json:"skill_count"`
	KnowledgeMountCount uint32       `json:"knowledge_mount_count"`
	EnvBundleCount      uint32       `json:"env_bundle_count"`
	Lifecycle           Lifecycle    `json:"lifecycle"`
}

func Summarize(spec Spec) (Summary, error) {
	normalized, err := NormalizeAndValidate(spec)
	if err != nil {
		return Summary{}, err
	}
	return Summary{
		Version:             normalized.Version,
		ModelBinding:        normalized.Runtime.ModelBinding,
		WorkerType:          normalized.Runtime.WorkerType,
		RuntimeImage:        normalized.Runtime.Image,
		Placement:           clonePlacement(normalized.Placement),
		Alias:               normalized.Metadata.Alias,
		RepositoryID:        cloneInt64Pointer(normalized.Workspace.RepositoryID),
		Branch:              normalized.Workspace.Branch,
		SkillCount:          uint32(len(normalized.Workspace.SkillIDs)),
		KnowledgeMountCount: uint32(len(normalized.Workspace.KnowledgeMounts)),
		EnvBundleCount:      uint32(len(normalized.Workspace.EnvBundleIDs)),
		Lifecycle:           normalized.Lifecycle,
	}, nil
}

func ValidateSummary(summary Summary) error {
	if summary.Version != VersionV1 {
		return fmt.Errorf("workerspec summary version %d is unsupported", summary.Version)
	}
	if err := validateModelBinding(summary.ModelBinding); err != nil {
		return fmt.Errorf("workerspec summary: %w", err)
	}
	if err := validateWorkerType(summary.WorkerType); err != nil {
		return err
	}
	if err := validateRuntimeImage(summary.RuntimeImage); err != nil {
		return err
	}
	if err := validatePlacement(summary.Placement); err != nil {
		return err
	}
	if utf8.RuneCountInString(summary.Alias) > maxAliasRunes {
		return fmt.Errorf("workerspec summary alias exceeds %d characters", maxAliasRunes)
	}
	switch {
	case summary.RepositoryID == nil && summary.Branch != "":
		return fmt.Errorf("workerspec summary repository is required when branch is set")
	case summary.RepositoryID != nil && *summary.RepositoryID <= 0:
		return fmt.Errorf("workerspec summary repository id must be positive")
	case summary.RepositoryID != nil && summary.Branch == "":
		return fmt.Errorf("workerspec summary branch is required with a repository")
	}
	if err := validateLifecycle(summary.Lifecycle); err != nil {
		return fmt.Errorf("workerspec summary: %w", err)
	}
	return nil
}
