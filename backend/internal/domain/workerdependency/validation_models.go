package workerdependency

import (
	"fmt"
	"regexp"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

var environmentTargetPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

func validateModels(document Document, models Models) error {
	if models.Primary != nil {
		if err := validateModel(document, *models.Primary); err != nil {
			return fmt.Errorf("primary model: %w", err)
		}
		if !containsModality(models.Primary.Modalities, airesource.ModalityChat) ||
			!containsCapability(
				models.Primary.Capabilities,
				airesource.CapabilityTextGeneration,
			) {
			return fmt.Errorf("primary model must support chat text generation")
		}
	}
	roles := make(map[slugkit.Slug]struct{}, len(models.Tools))
	targets := make(map[string]struct{}, len(models.Tools)*3)
	for _, tool := range models.Tools {
		if err := validateReference(
			document,
			tool.Binding,
			resource.KindToolBinding,
		); err != nil {
			return fmt.Errorf("tool model %q: %w", tool.Role, err)
		}
		if err := slugkit.Validate(tool.Role.String()); err != nil {
			return fmt.Errorf("tool model role: %w", err)
		}
		if _, exists := roles[tool.Role]; exists {
			return fmt.Errorf("duplicate tool model role %q", tool.Role)
		}
		roles[tool.Role] = struct{}{}
		if err := validateModel(document, tool.Model); err != nil {
			return fmt.Errorf("tool model %q: %w", tool.Role, err)
		}
		if !tool.Modality.Valid() || !tool.Capability.Valid() ||
			!containsModality(tool.Model.Modalities, tool.Modality) ||
			!containsCapability(tool.Model.Capabilities, tool.Capability) {
			return fmt.Errorf("tool model %q capability contract is invalid", tool.Role)
		}
		for _, target := range []string{
			tool.Environment.APIKeyTarget,
			tool.Environment.BaseURLTarget,
			tool.Environment.ModelIDTarget,
		} {
			if !environmentTargetPattern.MatchString(target) {
				return fmt.Errorf("tool model %q environment target is invalid", tool.Role)
			}
			if _, exists := targets[target]; exists {
				return fmt.Errorf("duplicate tool model environment target %q", target)
			}
			targets[target] = struct{}{}
		}
	}
	return nil
}

func validateModel(document Document, model Model) error {
	if err := validatePin(document, model.Pin, resource.KindModelBinding); err != nil {
		return err
	}
	switch {
	case model.ResourceRevision <= 0:
		return fmt.Errorf("model resource revision must be positive")
	case model.ConnectionID <= 0:
		return fmt.Errorf("model connection id must be positive")
	case model.ConnectionRevision <= 0:
		return fmt.Errorf("model connection revision must be positive")
	}
	if err := slugkit.Validate(model.ProviderKey.String()); err != nil {
		return fmt.Errorf("model provider key: %w", err)
	}
	if err := slugkit.Validate(model.ProtocolAdapter.String()); err != nil {
		return fmt.Errorf("model protocol adapter: %w", err)
	}
	if err := requireNormalized("model id", model.ModelID); err != nil {
		return err
	}
	if err := requireNormalized("model base url", model.BaseURL); err != nil {
		return err
	}
	if containsRawSecretText(model.BaseURL) || containsURLUserInfo(model.BaseURL) {
		return fmt.Errorf("model base url contains raw secret-like data")
	}
	if err := validateModalities(model.Modalities); err != nil {
		return err
	}
	return validateCapabilities(model.Capabilities)
}

func validateModalities(values []airesource.Modality) error {
	if len(values) == 0 {
		return fmt.Errorf("model modalities are required")
	}
	seen := make(map[airesource.Modality]struct{}, len(values))
	for _, value := range values {
		if !value.Valid() {
			return fmt.Errorf("model modality %q is invalid", value)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("duplicate model modality %q", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateCapabilities(values []airesource.Capability) error {
	if len(values) == 0 {
		return fmt.Errorf("model capabilities are required")
	}
	seen := make(map[airesource.Capability]struct{}, len(values))
	for _, value := range values {
		if !value.Valid() {
			return fmt.Errorf("model capability %q is invalid", value)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("duplicate model capability %q", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func containsModality(values []airesource.Modality, wanted airesource.Modality) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func containsCapability(
	values []airesource.Capability,
	wanted airesource.Capability,
) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
