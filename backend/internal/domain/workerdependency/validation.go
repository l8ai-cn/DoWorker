package workerdependency

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/google/uuid"
)

var (
	digestPattern     = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	definitionPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	commitPattern     = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)
)

func Validate(document Document) error {
	if document.Version != VersionV1 {
		return fmt.Errorf("worker dependency version %d is unsupported", document.Version)
	}
	if document.OrganizationID <= 0 {
		return fmt.Errorf("worker dependency organization id must be positive")
	}
	if err := slugkit.Validate(document.Namespace.String()); err != nil {
		return fmt.Errorf("worker dependency namespace: %w", err)
	}
	if err := validateWorker(document.Worker); err != nil {
		return err
	}
	if err := validateModels(document, document.Models); err != nil {
		return err
	}
	if err := validateAgentfileFieldOwnership(document); err != nil {
		return err
	}
	if err := validateRepository(document, document.Repository); err != nil {
		return err
	}
	if err := validateSkills(document, document.Skills); err != nil {
		return err
	}
	if err := validateKnowledgeBases(document, document.KnowledgeBases); err != nil {
		return err
	}
	if err := validateBundles(document, document.RuntimeBundles); err != nil {
		return err
	}
	if err := validateSecretReferences(
		document,
		document.RuntimeBundles,
		document.SecretReferences,
	); err != nil {
		return err
	}
	return validatePlacement(document, document.Placement)
}

func validateWorker(worker Worker) error {
	if err := slugkit.Validate(worker.WorkerType.String()); err != nil {
		return fmt.Errorf("worker dependency worker type: %w", err)
	}
	if err := slugkit.Validate(worker.AdapterID.String()); err != nil {
		return fmt.Errorf("worker dependency adapter id: %w", err)
	}
	if worker.SpecVersion != workerspec.VersionV1 {
		return fmt.Errorf("worker dependency workerspec version must be 1")
	}
	if err := validateDigest("worker dependency workerspec digest", worker.SpecDigest); err != nil {
		return err
	}
	if !definitionPattern.MatchString(worker.DefinitionHash) {
		return fmt.Errorf("worker dependency definition hash must be lowercase SHA-256 hex")
	}
	if err := validateWorkerEnvironmentPolicy(worker); err != nil {
		return err
	}
	if worker.AgentfileSource == "" {
		return fmt.Errorf("worker dependency agentfile source is required")
	}
	if err := validateTextDigest(
		"worker dependency agentfile source",
		worker.AgentfileSource,
		worker.AgentfileSourceDigest,
	); err != nil {
		return err
	}
	return validateAgentfileSyntax(worker.AgentfileSource)
}

func validateWorkerEnvironmentPolicy(worker Worker) error {
	modelFields, err := validateEnvironmentFields(
		"worker model-managed field",
		worker.ModelManagedFields,
	)
	if err != nil {
		return err
	}
	credentialFields, err := validateEnvironmentFields(
		"worker credential-bundle field",
		worker.CredentialBundleFields,
	)
	if err != nil {
		return err
	}
	for field := range modelFields {
		if _, exists := credentialFields[field]; exists {
			return fmt.Errorf("worker environment field %q has conflicting ownership", field)
		}
	}
	return nil
}

func validateEnvironmentFields(
	label string,
	fields []string,
) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if err := workerspec.ValidateConfigField(field); err != nil {
			return nil, fmt.Errorf("%s: %w", label, err)
		}
		if _, exists := seen[field]; exists {
			return nil, fmt.Errorf("duplicate %s %q", label, field)
		}
		seen[field] = struct{}{}
	}
	return seen, nil
}

func validatePin(
	document Document,
	pin ResourcePin,
	expectedKind string,
) error {
	if pin.DomainID <= 0 {
		return fmt.Errorf("%s dependency domain id must be positive", expectedKind)
	}
	return validateReference(document, pin.Reference, expectedKind)
}

func validateReference(
	document Document,
	reference orchestrationresource.Reference,
	expectedKind string,
) error {
	if reference.Kind != expectedKind {
		return fmt.Errorf(
			"%s dependency reference kind must be %q",
			expectedKind,
			expectedKind,
		)
	}
	if err := reference.ValidateResolved(document.Namespace.String()); err != nil {
		return fmt.Errorf("%s dependency reference: %w", expectedKind, err)
	}
	parsed, err := uuid.Parse(reference.UID)
	if err != nil {
		return fmt.Errorf("%s dependency reference uid must be a UUID", expectedKind)
	}
	if parsed.String() != reference.UID {
		return fmt.Errorf("%s dependency reference uid must be canonical", expectedKind)
	}
	return nil
}

func validateDigest(field, digest string) error {
	if !digestPattern.MatchString(digest) {
		return fmt.Errorf("%s must be an immutable SHA-256 digest", field)
	}
	return nil
}

func validateTextDigest(field, value, digest string) error {
	if err := validateDigest(field+" digest", digest); err != nil {
		return err
	}
	if TextDigest(value) != digest {
		return fmt.Errorf("%s digest does not match content", field)
	}
	return nil
}

func validateCommit(field, commit string) error {
	if !commitPattern.MatchString(commit) {
		return fmt.Errorf("%s must be a lowercase Git commit SHA", field)
	}
	return nil
}

func requireNormalized(field, value string) error {
	if value == "" || strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must be non-empty and normalized", field)
	}
	return nil
}
