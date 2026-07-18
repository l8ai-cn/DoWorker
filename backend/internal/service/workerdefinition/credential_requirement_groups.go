package workerdefinition

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type CredentialRequirementGroup struct {
	ID    string
	AnyOf []string
}

type credentialRequirementGroupDocument struct {
	ID    string   `json:"id"`
	AnyOf []string `json:"any_of"`
}

func validateCredentialRequirementGroups(
	rawGroups []json.RawMessage,
) error {
	_, err := decodeCredentialRequirementGroups(rawGroups)
	return err
}

func decodeCredentialRequirementGroups(
	rawGroups []json.RawMessage,
) ([]CredentialRequirementGroup, error) {
	groups := make([]CredentialRequirementGroup, 0, len(rawGroups))
	ids := make(map[string]struct{}, len(rawGroups))
	for _, raw := range rawGroups {
		var document credentialRequirementGroupDocument
		if err := decodeStrict(raw, &document); err != nil {
			return nil, err
		}
		if err := slugkit.Validate(document.ID); err != nil {
			return nil, fmt.Errorf("invalid credential requirement group id %q: %w", document.ID, err)
		}
		if len(document.AnyOf) < 2 {
			return nil, fmt.Errorf("credential requirement group %q must declare at least two targets", document.ID)
		}
		if _, exists := ids[document.ID]; exists {
			return nil, fmt.Errorf("duplicate credential requirement group id %q", document.ID)
		}
		targets := make(map[string]struct{}, len(document.AnyOf))
		for _, target := range document.AnyOf {
			if target == "" {
				return nil, fmt.Errorf("credential requirement group %q has an empty target", document.ID)
			}
			if _, exists := targets[target]; exists {
				return nil, fmt.Errorf("credential requirement group %q has duplicate target %q", document.ID, target)
			}
			targets[target] = struct{}{}
		}
		ids[document.ID] = struct{}{}
		groups = append(groups, CredentialRequirementGroup{
			ID: document.ID, AnyOf: append([]string{}, document.AnyOf...),
		})
	}
	return groups, nil
}

func validateCredentialRequirementGroupTargets(
	groups []CredentialRequirementGroup,
	bindings []CredentialBinding,
) error {
	byTarget := make(map[string]CredentialBinding, len(bindings))
	for _, binding := range bindings {
		byTarget[binding.Target.Name] = binding
	}
	usedTargets := make(map[string]string)
	for _, group := range groups {
		for _, target := range group.AnyOf {
			binding, exists := byTarget[target]
			if !exists {
				return fmt.Errorf("credential requirement group %q references undeclared target %q", group.ID, target)
			}
			if binding.Source.Kind != "credential_bundle" {
				return fmt.Errorf("credential requirement group %q target %q is not a credential bundle", group.ID, target)
			}
			if previous, exists := usedTargets[target]; exists {
				return fmt.Errorf("credential target %q belongs to groups %q and %q", target, previous, group.ID)
			}
			usedTargets[target] = group.ID
		}
	}
	return nil
}

func cloneCredentialRequirementGroups(
	groups []CredentialRequirementGroup,
) []CredentialRequirementGroup {
	cloned := append([]CredentialRequirementGroup{}, groups...)
	for index := range cloned {
		cloned[index].AnyOf = append([]string{}, cloned[index].AnyOf...)
	}
	return cloned
}
