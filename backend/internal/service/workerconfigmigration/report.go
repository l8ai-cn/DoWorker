package workerconfigmigration

type Report struct {
	SnapshotsScanned        int      `json:"snapshots_scanned"`
	SnapshotUpdates         int      `json:"snapshot_updates"`
	WorkerTemplateRevisions int      `json:"worker_template_revisions_scanned"`
	WorkerTemplateUpdates   int      `json:"worker_template_updates"`
	PendingLegacyPlans      int      `json:"pending_legacy_plans"`
	Blockers                []string `json:"blockers"`
}

func (report Report) Ready() bool {
	return len(report.Blockers) == 0
}

func (report Report) Clean() bool {
	return report.Ready() &&
		report.SnapshotUpdates == 0 &&
		report.WorkerTemplateUpdates == 0 &&
		report.PendingLegacyPlans == 0
}
