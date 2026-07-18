package agentpod

import "strconv"

func PendingCommandRunnerLockName(runnerID int64) string {
	return "pending-command-runner:" + strconv.FormatInt(runnerID, 10)
}
