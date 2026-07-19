package workerdependency

import (
	"fmt"

	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func validateRepository(document Document, repository *Repository) error {
	if repository == nil {
		return nil
	}
	if err := validatePin(document, repository.Pin, resource.KindRepository); err != nil {
		return err
	}
	if repository.HTTPCloneURL == "" && repository.SSHCloneURL == "" {
		return fmt.Errorf("repository requires an immutable clone endpoint")
	}
	if containsRawSecretText(repository.HTTPCloneURL) ||
		containsRawSecretText(repository.SSHCloneURL) ||
		containsURLUserInfo(repository.HTTPCloneURL) ||
		containsURLUserInfo(repository.SSHCloneURL) {
		return fmt.Errorf("repository clone endpoint contains raw secret-like data")
	}
	if err := requireNormalized("repository branch", repository.Branch); err != nil {
		return err
	}
	if err := validateCommit("repository commit", repository.CommitSHA); err != nil {
		return err
	}
	if err := validateRepositoryCredential(*repository); err != nil {
		return err
	}
	if repository.PreparationScript == "" {
		if repository.PreparationScriptDigest != "" ||
			repository.PreparationTimeoutSeconds != 0 {
			return fmt.Errorf("empty repository preparation must not carry metadata")
		}
		return nil
	}
	if containsRawSecretText(repository.PreparationScript) {
		return fmt.Errorf("repository preparation script contains raw secret-like data")
	}
	if repository.PreparationTimeoutSeconds == 0 {
		return fmt.Errorf("repository preparation timeout must be positive")
	}
	return validateTextDigest(
		"repository preparation script",
		repository.PreparationScript,
		repository.PreparationScriptDigest,
	)
}

func validateRepositoryCredential(repository Repository) error {
	credential := repository.Credential
	switch credential.Type {
	case RepositoryCredentialTypeNone:
		if credential.CredentialID != nil || credential.OwnerUserID != 0 {
			return fmt.Errorf("unauthenticated repository clone must not carry a credential id")
		}
		if repository.HTTPCloneURL == "" {
			return fmt.Errorf("unauthenticated repository clone requires an HTTP endpoint")
		}
	case user.CredentialTypeRunnerLocal:
		return fmt.Errorf("runner-local repository clone requires an exact Runner secret reference")
	case user.CredentialTypeOAuth, user.CredentialTypePAT:
		if credential.CredentialID == nil || *credential.CredentialID <= 0 {
			return fmt.Errorf("repository credential id must be positive")
		}
		if credential.OwnerUserID <= 0 {
			return fmt.Errorf("repository credential owner user id must be positive")
		}
		if repository.HTTPCloneURL == "" {
			return fmt.Errorf("token repository credential requires an HTTP endpoint")
		}
	case user.CredentialTypeSSHKey:
		if credential.CredentialID == nil || *credential.CredentialID <= 0 {
			return fmt.Errorf("repository credential id must be positive")
		}
		if credential.OwnerUserID <= 0 {
			return fmt.Errorf("repository credential owner user id must be positive")
		}
		if repository.SSHCloneURL == "" {
			return fmt.Errorf("SSH repository credential requires an SSH endpoint")
		}
	default:
		return fmt.Errorf("repository credential type %q is invalid", credential.Type)
	}
	return nil
}

func validateSkills(document Document, skills []Skill) error {
	seen := make(map[string]struct{}, len(skills))
	domainIDs := make(map[int64]struct{}, len(skills))
	for _, skill := range skills {
		if err := validatePin(document, skill.Pin, resource.KindSkill); err != nil {
			return err
		}
		key := referenceKey(skill.Pin)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate Skill dependency")
		}
		seen[key] = struct{}{}
		if _, exists := domainIDs[skill.Pin.DomainID]; exists {
			return fmt.Errorf("duplicate Skill domain id %d", skill.Pin.DomainID)
		}
		domainIDs[skill.Pin.DomainID] = struct{}{}
		if err := slugkit.Validate(skill.Slug.String()); err != nil {
			return fmt.Errorf("Skill slug: %w", err)
		}
		if skill.Version <= 0 {
			return fmt.Errorf("Skill version must be positive")
		}
		if err := validateDigest("Skill content digest", skill.ContentDigest); err != nil {
			return err
		}
		if err := requireNormalized("Skill storage key", skill.StorageKey); err != nil {
			return err
		}
		if skill.PackageSize < 0 {
			return fmt.Errorf("Skill package size must not be negative")
		}
	}
	return nil
}

func validateKnowledgeBases(document Document, items []KnowledgeBase) error {
	seen := make(map[string]struct{}, len(items))
	domainIDs := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if err := validatePin(
			document,
			item.Pin,
			resource.KindKnowledgeBase,
		); err != nil {
			return err
		}
		key := referenceKey(item.Pin)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate KnowledgeBase dependency")
		}
		seen[key] = struct{}{}
		if _, exists := domainIDs[item.Pin.DomainID]; exists {
			return fmt.Errorf(
				"duplicate KnowledgeBase domain id %d",
				item.Pin.DomainID,
			)
		}
		domainIDs[item.Pin.DomainID] = struct{}{}
		if err := slugkit.Validate(item.Slug.String()); err != nil {
			return fmt.Errorf("KnowledgeBase slug: %w", err)
		}
		if err := requireNormalized(
			"KnowledgeBase clone url",
			item.HTTPCloneURL,
		); err != nil {
			return err
		}
		if containsRawSecretText(item.HTTPCloneURL) ||
			containsURLUserInfo(item.HTTPCloneURL) {
			return fmt.Errorf("KnowledgeBase clone url contains raw secret-like data")
		}
		if err := requireNormalized("KnowledgeBase branch", item.Branch); err != nil {
			return err
		}
		if err := validateCommit("KnowledgeBase commit", item.CommitSHA); err != nil {
			return err
		}
		switch item.Mode {
		case workerspec.KnowledgeMountReadOnly, workerspec.KnowledgeMountReadWrite:
		default:
			return fmt.Errorf("KnowledgeBase mount mode %q is invalid", item.Mode)
		}
	}
	return nil
}
