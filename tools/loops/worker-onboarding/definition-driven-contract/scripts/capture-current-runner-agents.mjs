import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../../../..");
const defaultOutput = path.join(
  root,
  "tools/loops/worker-onboarding/definition-driven-contract/evidence/current-runner-agents.json",
);
const query = `
SELECT COALESCE(
  jsonb_agg(
    jsonb_build_object(
      'node_id', node_id,
      'status', status,
      'available_agents', available_agents,
      'agent_versions', agent_versions
    )
    ORDER BY node_id
  ),
  '[]'::jsonb
)
FROM runners
WHERE node_id LIKE 'dev-runner-%';
`;

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}

function main() {
  const output = outputPath(process.argv.slice(2));
  const project = readComposeProjectName();
  const result = spawnSync(
    "docker",
    [
      "exec",
      `${project}-postgres-1`,
      "psql",
      "-U",
      "agentsmesh",
      "-d",
      "agentsmesh",
      "-Atc",
      query,
    ],
    { encoding: "utf8" },
  );
  if (result.status !== 0) throw new Error(result.stderr.trim() || "runner query failed");
  const runners = JSON.parse(result.stdout.trim());
  if (!Array.isArray(runners)) throw new Error("runner query returned a non-array");
  const document = {
    schema_version: 1,
    observed_at: new Date().toISOString(),
    source: "development runners table",
    runners,
  };
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, JSON.stringify(document, null, 2) + "\n");
  console.log(`wrote current Runner agents: ${output}`);
}

function outputPath(argumentsList) {
  if (argumentsList.length === 0) return defaultOutput;
  if (argumentsList.length === 2 && argumentsList[0] === "--output") {
    return path.resolve(argumentsList[1]);
  }
  throw new Error("usage: node capture-current-runner-agents.mjs [--output <path>]");
}

function readComposeProjectName() {
  const environment = fs.readFileSync(path.join(root, "deploy/dev/.env"), "utf8");
  const match = environment.match(/^COMPOSE_PROJECT_NAME=([a-z0-9][a-z0-9_-]*)$/m);
  if (!match) throw new Error("deploy/dev/.env must define COMPOSE_PROJECT_NAME");
  return match[1];
}
