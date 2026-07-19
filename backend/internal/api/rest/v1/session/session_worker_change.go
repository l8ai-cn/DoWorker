package sessionapi

import "strings"

func hasSessionWorkerConfigChange(
	modelResourceID *int64,
	workerSpec *sessionWorkerSpecBody,
	automationLevel string,
) bool {
	return modelResourceID != nil ||
		workerSpec != nil ||
		strings.TrimSpace(automationLevel) != ""
}

func rejectSameAgentWorkerConfigChangeMessage() string {
	return "same-agent operation cannot change worker configuration"
}
