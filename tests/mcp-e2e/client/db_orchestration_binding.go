package client

import (
	"context"
	"fmt"
)

type OrchestrationDomainBinding struct {
	Status               string
	ResourceID           int64
	ResourceRevision     int64
	WorkerSpecSnapshotID int64
}

func (d *DB) GetWorkflowOrchestrationBinding(
	ctx context.Context,
	orgSlug, workflowSlug string,
) (OrchestrationDomainBinding, error) {
	var binding OrchestrationDomainBinding
	err := d.conn.QueryRowContext(ctx, `
SELECT workflow.status,
       workflow.orchestration_resource_id,
       workflow.orchestration_resource_revision,
       workflow.worker_spec_snapshot_id
FROM workflows workflow
JOIN organizations organization
  ON organization.id = workflow.organization_id
WHERE organization.slug = $1 AND workflow.slug = $2`,
		orgSlug,
		workflowSlug,
	).Scan(
		&binding.Status,
		&binding.ResourceID,
		&binding.ResourceRevision,
		&binding.WorkerSpecSnapshotID,
	)
	if err != nil {
		return OrchestrationDomainBinding{}, fmt.Errorf(
			"load workflow orchestration binding %s: %w",
			workflowSlug,
			err,
		)
	}
	return binding, nil
}

func (d *DB) GetWorkerLaunchOrchestrationBinding(
	ctx context.Context,
	podKey string,
) (OrchestrationDomainBinding, error) {
	var binding OrchestrationDomainBinding
	err := d.conn.QueryRowContext(ctx, `
SELECT pod.status,
       launch.resource_id,
       launch.resource_revision,
       launch.worker_spec_snapshot_id
FROM orchestration_worker_launches launch
JOIN pods pod ON pod.id = launch.pod_id
WHERE pod.pod_key = $1`,
		podKey,
	).Scan(
		&binding.Status,
		&binding.ResourceID,
		&binding.ResourceRevision,
		&binding.WorkerSpecSnapshotID,
	)
	if err != nil {
		return OrchestrationDomainBinding{}, fmt.Errorf(
			"load worker launch orchestration binding %s: %w",
			podKey,
			err,
		)
	}
	return binding, nil
}
