import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../../../..");
const defaultOutput = path.join(
  root,
  "tools/loops/worker-onboarding/definition-driven-contract/evidence/current-live-create-options.json",
);

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}

async function main() {
  const output = outputPath(process.argv.slice(2));
  const baseURL = requiredEnvironment("WORKER_EVIDENCE_API_BASE").replace(/\/+$/, "");
  const username = requiredEnvironment("WORKER_EVIDENCE_USERNAME");
  const password = requiredEnvironment("WORKER_EVIDENCE_PASSWORD");
  const orgSlug = requiredEnvironment("WORKER_EVIDENCE_ORG_SLUG");
  const token = await login(baseURL, username, password);
  const response = await requestOptions(baseURL, token, orgSlug);
  const document = {
    schema_version: 1,
    observed_at: new Date().toISOString(),
    endpoint: "POST /proto.pod.v1.PodService/ListWorkerCreateOptions",
    scope: orgSlug,
    safety_boundary: {
      pod_create_rpc: false,
      provider_request: false,
      credential_read_or_decrypt: false,
    },
    revision: response.revision,
    worker_types: response.workerTypes.map(workerTypeEvidence).sort(compareSlug),
    runtime_images: response.runtimeImages.map(runtimeImageEvidence).sort(compareSlug),
  };
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, JSON.stringify(document, null, 2) + "\n");
  console.log(`wrote current Worker create options: ${output}`);
}

async function login(baseURL, username, password) {
  const response = await fetch(`${baseURL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!response.ok) throw new Error(`development login failed: ${response.status}`);
  const payload = await response.json();
  if (typeof payload.token !== "string" || payload.token === "") {
    throw new Error("development login returned no access token");
  }
  return payload.token;
}

async function requestOptions(baseURL, token, orgSlug) {
  const response = await fetch(
    `${baseURL}/proto.pod.v1.PodService/ListWorkerCreateOptions`,
    {
      method: "POST",
      headers: {
        Accept: "application/json",
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ orgSlug }),
    },
  );
  if (!response.ok) throw new Error(`list Worker create options failed: ${response.status}`);
  return response.json();
}

function workerTypeEvidence(value) {
  return {
    slug: value.slug,
    selectable: value.selectable === true,
    blocking_reason: value.blockingReason ?? "",
    supported_interaction_modes: value.supportedInteractionModes ?? [],
    model_protocol_adapters: value.modelProtocolAdapters ?? [],
    config_schema: parseConfigSchema(value.configSchemaJson),
  };
}

function parseConfigSchema(raw) {
  if (typeof raw !== "string") {
    throw new Error("worker create option is missing configSchemaJson");
  }
  const value = JSON.parse(raw);
  if (value === null || Array.isArray(value) || typeof value !== "object") {
    throw new Error("worker create option config schema must be an object");
  }
  return value;
}

function runtimeImageEvidence(value) {
  return {
    slug: value.slug,
    selectable: value.selectable === true,
    blocking_reason: value.blockingReason ?? "",
    worker_type_slugs: value.workerTypeSlugs ?? [],
  };
}

function compareSlug(left, right) {
  return left.slug.localeCompare(right.slug);
}

function outputPath(argumentsList) {
  if (argumentsList.length === 0) return defaultOutput;
  if (argumentsList.length === 2 && argumentsList[0] === "--output") {
    return path.resolve(argumentsList[1]);
  }
  throw new Error("usage: node capture-current-create-options.mjs [--output <path>]");
}

function requiredEnvironment(name) {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
}
