import type { ApiFixture } from "../fixtures/api.fixture";
import type { DbFixture } from "../fixtures/db.fixture";
import { TEST_ORG_SLUG } from "./env";
import { terminateRegisteredE2EPods } from "./pod-cleanup";
import { resetResourceWorkflowFixture } from "./resource-workflow-fixture";
import { RESOURCE_WORKFLOW_SLUG } from "./resource-workflow-manifest";

export async function cleanupResourceWorkflowFixture(
  db: DbFixture,
  api: ApiFixture,
): Promise<void> {
  const runIDs = activeRunIDs(db);
  if (runIDs.length > 0) {
    const client = await api.connect();
    for (const runID of runIDs) {
      await client.workflow.cancelWorkflowRun({
        orgSlug: TEST_ORG_SLUG,
        workflowSlug: RESOURCE_WORKFLOW_SLUG,
        runId: BigInt(runID),
      });
    }
  }
  await terminateRegisteredE2EPods();
  resetResourceWorkflowFixture(db);
}

function activeRunIDs(db: DbFixture): string[] {
  const rows = db.queryValue(`
    SELECT run.id, COALESCE(run.pod_key, '')
    FROM workflow_runs run
    JOIN workflows workflow ON workflow.id = run.workflow_id
    WHERE workflow.organization_id = (
      SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
    ) AND workflow.slug = '${RESOURCE_WORKFLOW_SLUG}'
      AND run.status IN ('pending', 'running')
    ORDER BY run.id;
  `);
  if (!rows) return [];
  return rows.split("\n").map((row) => {
    const [runID, podKey] = row.split("|", 2);
    if (!/^[1-9][0-9]*$/.test(runID ?? "") || podKey === undefined) {
      throw new Error(`invalid active resource Workflow run: ${row}`);
    }
    return runID;
  });
}
