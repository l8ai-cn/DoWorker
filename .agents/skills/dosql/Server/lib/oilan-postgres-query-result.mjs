const OILAN_TARGETING_LINE = [
  "[TARGETING] Server: gw-oilan-node",
  "(https://doops.l8ai.cn -> doops-oilan/oilan-node),",
  "Use: doops-oilan/oilan-node via gateway",
].join(" ");

export const OILAN_DOOPS_ROUTE = Object.freeze({
  targetName: "gw-oilan-node",
  gateway: "https://doops.l8ai.cn",
  cluster: "doops-oilan",
  instance: "oilan-node",
});

export function parseOilanPostgresDoopsResult(queryName, stdout) {
  const rows = parseOilanDoopsRows(stdout);
  if (rows.length !== 1) {
    throw new Error("Oilan PostgreSQL query must return exactly one result row");
  }
  return {
    doopsRoute: OILAN_DOOPS_ROUTE,
    result: parseOilanPostgresQueryResult(queryName, rows[0]),
  };
}

export function parseOilanDoopsRows(stdout) {
  const lines = resultLines(stdout);
  if (lines[0] !== OILAN_TARGETING_LINE) {
    throw new Error("DoOps did not prove the canonical Oilan target route");
  }
  return lines.slice(1);
}

export function parseOilanPostgresQueryResult(queryName, line) {
  const fields = line.split("|");

  if (queryName === "asset-probe") {
    requireFieldCount(fields, 3, queryName);
    const [databaseName, serverVersionNum, schemaMigrationsPresent] = fields;
    if (databaseName !== "agentsmesh") {
      throw new Error("Oilan PostgreSQL probe returned the wrong database");
    }
    if (!/^[1-9][0-9]{4,5}$/.test(serverVersionNum)) {
      throw new Error("Oilan PostgreSQL probe returned an invalid server version");
    }
    if (schemaMigrationsPresent !== "t") {
      throw new Error("Oilan PostgreSQL schema_migrations table is missing");
    }
    return {
      databaseName,
      serverVersionNum,
      schemaMigrationsPresent: true,
    };
  }

  if (queryName === "migration-version") {
    requireFieldCount(fields, 2, queryName);
    const [versionText, dirtyText] = fields;
    if (!/^[0-9]+$/.test(versionText)) {
      throw new Error("Oilan PostgreSQL migration version is invalid");
    }
    if (dirtyText !== "t" && dirtyText !== "f") {
      throw new Error("Oilan PostgreSQL migration dirty state is invalid");
    }
    return {
      version: Number(versionText),
      dirty: dirtyText === "t",
    };
  }

  throw new Error(`unsupported Oilan PostgreSQL query result: ${queryName}`);
}

export function assertOilanPostgresQueryResult(queryName, result) {
  if (!result || typeof result !== "object" || Array.isArray(result)) {
    throw new Error("Oilan PostgreSQL result must be an object");
  }
  if (queryName === "asset-probe") {
    const parsed = parseOilanPostgresQueryResult(
      queryName,
      `${result.databaseName}|${result.serverVersionNum}|${result.schemaMigrationsPresent ? "t" : "f"}`,
    );
    if (Object.keys(result).sort().join(",") !== Object.keys(parsed).sort().join(",")) {
      throw new Error("Oilan PostgreSQL probe result has unexpected fields");
    }
    return true;
  }
  if (queryName === "migration-version") {
    const dirty = result.dirty === true ? "t" : result.dirty === false ? "f" : "";
    const parsed = parseOilanPostgresQueryResult(queryName, `${result.version}|${dirty}`);
    if (Object.keys(result).sort().join(",") !== Object.keys(parsed).sort().join(",")) {
      throw new Error("Oilan PostgreSQL migration result has unexpected fields");
    }
    return true;
  }
  throw new Error(`unsupported Oilan PostgreSQL query result: ${queryName}`);
}

export function assertOilanDoopsRoute(route) {
  for (const [key, value] of Object.entries(OILAN_DOOPS_ROUTE)) {
    if (route?.[key] !== value) {
      throw new Error("DoOps route does not match the canonical Oilan target");
    }
  }
  if (Object.keys(route).sort().join(",") !== Object.keys(OILAN_DOOPS_ROUTE).sort().join(",")) {
    throw new Error("DoOps route has unexpected fields");
  }
  return true;
}

function resultLines(stdout) {
  return String(stdout ?? "")
    .replace(/\u001b\[[0-9;]*m/g, "")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
}

function requireFieldCount(fields, expected, queryName) {
  if (fields.length !== expected) {
    throw new Error(`Oilan PostgreSQL ${queryName} result shape is invalid`);
  }
}
