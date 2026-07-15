package agentpod

type PodResourceBindings struct {
	VirtualAPIKeyID             *int64 `gorm:"column:virtual_api_key_id" json:"virtual_api_key_id,omitempty"`
	ModelResourceID             *int64 `gorm:"column:model_resource_id" json:"model_resource_id,omitempty"`
	WorkerSpecSnapshotID        *int64 `gorm:"column:worker_spec_snapshot_id" json:"worker_spec_snapshot_id,omitempty"`
	OrchestrationWorkerLaunchID *int64 `gorm:"column:orchestration_worker_launch_id" json:"orchestration_worker_launch_id,omitempty"`
}
