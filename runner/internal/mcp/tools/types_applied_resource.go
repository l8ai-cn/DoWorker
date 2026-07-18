package tools

import "fmt"

type AppliedResourceSummary struct {
	Kind                 string `json:"kind"`
	Name                 string `json:"name"`
	UID                  string `json:"uid"`
	Revision             int64  `json:"revision"`
	WorkerSpecSnapshotID int64  `json:"worker_spec_snapshot_id"`
}

func (resource *AppliedResourceSummary) FormatText() string {
	if resource == nil {
		return ""
	}
	text := fmt.Sprintf(
		"Resource: %s/%s@r%d",
		resource.Kind,
		resource.Name,
		resource.Revision,
	)
	if resource.WorkerSpecSnapshotID > 0 {
		text += fmt.Sprintf(
			" | Snapshot: %d",
			resource.WorkerSpecSnapshotID,
		)
	}
	return text
}
