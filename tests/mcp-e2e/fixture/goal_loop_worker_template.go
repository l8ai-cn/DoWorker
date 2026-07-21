package fixture

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/l8ai-cn/agentcloud/tests/mcp-e2e/client"
	"github.com/google/uuid"
)

func seedGoalLoopWorkerTemplate(
	ctx context.Context,
	db *client.DB,
	env *Env,
) (string, error) {
	var organizationID, actorID int64
	if err := db.QueryRow(ctx, `
SELECT organization.id, member.user_id
FROM organizations organization
JOIN organization_members member
  ON member.organization_id = organization.id
JOIN users actor ON actor.id = member.user_id
WHERE organization.slug = $1 AND actor.username = $2`,
		env.DevOrgSlug,
		env.DevUser,
	).Scan(&organizationID, &actorID); err != nil {
		return "", fmt.Errorf("resolve goal loop resource scope: %w", err)
	}

	name := uniqueAlias("e2e-loop-template")
	resourceUID := uuid.NewString()
	spec := staleWorkerTemplateSpec()
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("encode worker template spec: %w", err)
	}
	manifestJSON, err := json.Marshal(map[string]any{
		"apiVersion": "agentcloud.io/v1alpha1",
		"kind":       "WorkerTemplate",
		"metadata": map[string]any{
			"name": name, "namespace": env.DevOrgSlug,
			"displayName": name, "uid": resourceUID,
			"resourceVersion": "1", "generation": 1,
		},
		"spec":   spec,
		"status": map[string]any{},
	})
	if err != nil {
		return "", fmt.Errorf("encode worker template manifest: %w", err)
	}
	digest := sha256.Sum256(manifestJSON)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin worker template seed: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var snapshotID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO worker_spec_snapshots
  (organization_id, version, spec_json, summary_json)
VALUES ($1, 1, $2::jsonb, $3::jsonb)
RETURNING id`,
		organizationID,
		goalLoopSpecJSON,
		goalLoopSummaryJSON,
	).Scan(&snapshotID); err != nil {
		return "", fmt.Errorf("insert stale worker snapshot: %w", err)
	}

	var resourceID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO orchestration_resources
  (organization_id, uid, api_version, kind, namespace, name,
   display_name, generation, resource_version, active_revision,
   created_by_id, updated_by_id)
VALUES ($1, $2, 'agentcloud.io/v1alpha1', 'WorkerTemplate', $3, $4,
        $4, 1, 1, 1, $5, $5)
RETURNING id`,
		organizationID,
		resourceUID,
		env.DevOrgSlug,
		name,
		actorID,
	).Scan(&resourceID); err != nil {
		return "", fmt.Errorf("insert worker template resource: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO orchestration_resource_revisions
  (organization_id, resource_id, revision, generation, resource_version,
   canonical_manifest, canonical_spec, resolved_refs, digest,
   worker_spec_snapshot_id, actor_id)
VALUES ($1, $2, 1, 1, 1, $3::jsonb, $4::jsonb, '[]'::jsonb,
        $5, $6, $7)`,
		organizationID,
		resourceID,
		manifestJSON,
		specJSON,
		fmt.Sprintf("sha256:%x", digest),
		snapshotID,
		actorID,
	); err != nil {
		return "", fmt.Errorf("insert worker template revision: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit worker template seed: %w", err)
	}
	return name, nil
}

func staleWorkerTemplateSpec() map[string]any {
	return map[string]any{
		"optionsRevision": "e2e-stale-runtime",
		"workerType":      "e2e-echo",
		"toolRefs":        map[string]any{},
		"runtime": map[string]any{
			"runtimeImageId": 1, "placementPolicy": "automatic",
			"computeTargetRef": map[string]any{
				"kind": "ComputeTarget", "name": "organization-runner-pool",
			},
			"deploymentMode": "pooled",
			"resourceProfileRef": map[string]any{
				"kind": "ResourceProfile", "name": "standard",
			},
		},
		"typeConfig": map[string]any{
			"schemaVersion": 1, "values": map[string]any{"scenario": "echo"},
			"secretRefs": map[string]any{}, "interactionMode": "acp",
			"automationLevel": "autonomous",
		},
		"workspace": map[string]any{
			"branch": "", "skillRefs": []any{}, "knowledgeMounts": []any{},
			"environmentBundleRefs":  []any{},
			"configDocumentBindings": []any{}, "instructions": "",
		},
		"lifecycle": map[string]any{
			"terminationPolicy": "manual", "idleTimeoutMinutes": 0,
		},
		"metadata": map[string]any{"alias": "e2e-goal-loop"},
	}
}
