package coordinator

import (
	"strings"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
)

func coordinatorSnapshotID(project *coordinatordom.Project) (*int64, error) {
	if project == nil ||
		project.WorkerSpecSnapshotID == nil ||
		*project.WorkerSpecSnapshotID <= 0 {
		return nil, ErrCoordinatorWorkerSpecSnapshotRequired
	}
	return project.WorkerSpecSnapshotID, nil
}

func buildTaskPrompt(repo string, task ExternalTask) string {
	var lines []string
	prompt := strings.TrimSpace(task.Title + "\n\n" + task.Description)
	lines = append(lines, prompt)
	if repo != "" {
		lines = append(lines, "Repository: "+repo)
	}
	return strings.Join(lines, "\n")
}
