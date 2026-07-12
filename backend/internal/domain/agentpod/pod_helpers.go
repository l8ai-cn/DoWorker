package agentpod

func ActiveStatuses() []string {
	return []string{StatusQueued, StatusInitializing, StatusRunning, StatusPaused, StatusDisconnected}
}

func TerminalStatuses() []string {
	return []string{StatusTerminated, StatusOrphaned, StatusError}
}

func IsPodStatusActive(status string) bool {
	return status == StatusQueued ||
		status == StatusRunning ||
		status == StatusInitializing ||
		status == StatusPaused ||
		status == StatusDisconnected
}

func IsPodStatusRelayConnectable(status string) bool {
	return status == StatusRunning ||
		status == StatusPaused ||
		status == StatusDisconnected
}

func IsPodStatusTerminal(status string) bool {
	return status == StatusTerminated ||
		status == StatusOrphaned ||
		status == StatusError
}

func IsPodStatusFinished(status string) bool {
	return status == StatusCompleted || IsPodStatusTerminal(status)
}
