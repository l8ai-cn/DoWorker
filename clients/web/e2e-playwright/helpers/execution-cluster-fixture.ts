import type { DbFixture } from "../fixtures/db.fixture";
import { TEST_ORG_SLUG } from "./env";

export function getTestExecutionClusterId(db: DbFixture): bigint {
  const value = db.queryValue(`
    SELECT id
    FROM execution_clusters
    WHERE organization_id = (
      SELECT id FROM organizations WHERE slug = '${TEST_ORG_SLUG}'
    )
    ORDER BY id
    LIMIT 1
  `);
  if (!value) {
    throw new Error(`No execution cluster exists for ${TEST_ORG_SLUG}`);
  }
  return BigInt(value);
}
