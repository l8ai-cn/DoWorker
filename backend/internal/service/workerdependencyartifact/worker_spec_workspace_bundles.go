package workerdependencyartifact

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func sameRuntimeBundles(
	expected []workerspec.RuntimeEnvBundleID,
	bundles []workerdependency.RuntimeBundle,
) bool {
	actual := make([]int64, 0, len(bundles))
	seen := make(map[int64]struct{}, len(bundles))
	for _, bundle := range bundles {
		if bundle.ConfigDocument == nil {
			if _, exists := seen[bundle.Pin.DomainID]; exists {
				return false
			}
			seen[bundle.Pin.DomainID] = struct{}{}
			actual = append(actual, bundle.Pin.DomainID)
		}
	}
	if len(expected) != len(actual) {
		return false
	}
	for index := range expected {
		if int64(expected[index]) != actual[index] {
			return false
		}
	}
	return true
}

func sameConfigDocuments(
	expected []workerspec.ConfigDocumentBinding,
	bundles []workerdependency.RuntimeBundle,
) bool {
	actual := make(map[string]int64, len(expected))
	for _, bundle := range bundles {
		if bundle.ConfigDocument != nil {
			if _, exists := actual[bundle.ConfigDocument.ID]; exists {
				return false
			}
			actual[bundle.ConfigDocument.ID] = bundle.Pin.DomainID
		}
	}
	if len(expected) != len(actual) {
		return false
	}
	for _, binding := range expected {
		if actual[binding.DocumentID] != binding.ConfigBundleID {
			return false
		}
	}
	return true
}

func sameSecretReferences(
	expected map[string]workerspec.SecretReference,
	references []workerdependency.SecretReference,
) bool {
	if len(expected) != len(references) {
		return false
	}
	actual := make(map[string]int64, len(references))
	for _, reference := range references {
		if _, exists := actual[reference.Field]; exists {
			return false
		}
		actual[reference.Field] = reference.Pin.DomainID
	}
	for field, reference := range expected {
		if reference.Kind.String() != "env-bundle" ||
			actual[field] != reference.ID {
			return false
		}
	}
	return true
}
