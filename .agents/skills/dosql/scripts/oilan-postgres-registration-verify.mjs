#!/usr/bin/env node
import { verifyOilanPostgresGatewayAudit } from "../Server/lib/oilan-postgres-gateway-audit.mjs";
import { loadOilanPostgresRegistration } from "../Server/lib/oilan-postgres-doops-registration.mjs";

const registration = await loadOilanPostgresRegistration();
const result = await verifyOilanPostgresGatewayAudit(registration);

console.log(JSON.stringify({
  status: "corroborated",
  releaseAuthority: false,
  blocker: "DoOps Gateway audit does not retain an immutable full-command digest",
  databaseAssetId: result.databaseAssetId,
  registrationStatus: result.registrationStatus,
  assetProbe: result.assetProbe,
  migrationState: result.migrationState,
}, null, 2));
