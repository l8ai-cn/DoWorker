package airesource

import domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"

func resourceRuntimeChanged(current, candidate *domain.ModelResource) bool {
	return current.ModelID != candidate.ModelID ||
		!sameUnorderedValues(current.Modalities, candidate.Modalities) ||
		!sameUnorderedValues(current.Capabilities, candidate.Capabilities)
}

func sameUnorderedValues[T comparable](left, right []T) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[T]int, len(left))
	for _, value := range left {
		counts[value]++
	}
	for _, value := range right {
		counts[value]--
		if counts[value] < 0 {
			return false
		}
	}
	return true
}
