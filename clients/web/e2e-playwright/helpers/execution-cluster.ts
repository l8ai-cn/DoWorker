import type { DbFixture } from "../fixtures/db.fixture";

export function getOnlineExecutionClusterId(
  db: DbFixture,
  organizationSlug: string,
): bigint {
  const clusterId = db.queryValue(`
    SELECT cluster.id
    FROM execution_clusters AS cluster
    JOIN organizations AS organization
      ON organization.id = cluster.organization_id
    WHERE organization.slug = '${organizationSlug}'
      AND cluster.slug = 'online'
    LIMIT 1
  `);
  if (!clusterId) {
    throw new Error(`online execution cluster not found for ${organizationSlug}`);
  }
  return BigInt(clusterId);
}
