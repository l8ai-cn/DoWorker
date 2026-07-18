import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../../../..");
const workerTypesRoot = path.join(root, "config", "worker-types");
const loopRoot = path.join(
  root,
  "tools/loops/worker-onboarding/definition-driven-contract",
);
const matrixPath = path.join(loopRoot, "evidence/current-worker-evidence-matrix.json");
const defaultOutput = path.join(
  loopRoot,
  "evidence/worker-lifecycle-templates-2026-07-16.md",
);

const runtimeDetails = {
  aider: {
    build: "AGENT_RUNTIME=aider installs aider-chat with pip.",
    runner: "PTY only; the Runner registers the aider process and usage parser.",
  },
  "claude-code": {
    build: "AGENT_RUNTIME=claude-code installs @anthropic-ai/claude-code.",
    runner: "Custom claude-stream-json transport with Claude streaming and control handling.",
  },
  "codex-cli": {
    build: "AGENT_RUNTIME=codex-cli installs @openai/codex.",
    runner: "Custom codex-app-server transport; isolated CODEX_HOME and auth/config materialization.",
  },
  "cursor-cli": {
    build: "AGENT_RUNTIME=cursor-cli installs Cursor and exposes the agent binary.",
    runner: "Standard ACP transport registered as cursor-acp.",
  },
  "do-agent": {
    build: "AGENT_RUNTIME=do-agent stages the real do-agent sidecar binary.",
    runner: "Custom do-agent-acp transport with allow/restricted permission modes.",
  },
  "gemini-cli": {
    build: "AGENT_RUNTIME=gemini-cli installs @google/gemini-cli.",
    runner: "Standard ACP transport registered as gemini-acp; model launch argument is required.",
  },
  "grok-build": {
    build: "AGENT_RUNTIME=grok-build installs @xai-official/grok.",
    runner: "ACP transport performs xai.api_key headless authentication during initialize.",
  },
  hermes: {
    build: "AGENT_RUNTIME=hermes installs hermes-agent on the Python runtime base.",
    runner: "PTY only; HERMES_HOME is isolated per pod.",
  },
  loopal: {
    build: "AGENT_RUNTIME=loopal requires a real LOOPAL_BINARY and rejects E2E mock artifacts.",
    runner: "ACP transport forwards Loopal control-panel events and control requests.",
  },
  "minimax-cli": {
    build: "AGENT_RUNTIME=minimax-cli installs mmx-cli behind the MMX config wrapper.",
    runner: "PTY only; the wrapper writes MINIMAX_API_KEY to MMX_CONFIG_DIR.",
  },
  openclaw: {
    build: "AGENT_RUNTIME=openclaw installs OpenClaw with the pinned Node runtime.",
    runner: "PTY only; OPENCLAW_HOME config is merged and receives OpenAI provider settings.",
  },
  opencode: {
    build: "AGENT_RUNTIME=opencode installs opencode-ai.",
    runner: "Standard ACP transport registered as opencode-acp.",
  },
  "seedance-expert": {
    build: "Uses AGENT_RUNTIME=do-agent and the real do-agent sidecar binary.",
    runner: "Uses do-agent-acp plus a required Seedance video tool-model environment.",
  },
};

if (process.argv[1] === fileURLToPath(import.meta.url)) main();

function main() {
  const catalog = readJson(path.join(workerTypesRoot, "catalog.json"));
  const matrix = readJson(matrixPath);
  const output = outputPath(process.argv.slice(2));
  const evidence = new Map(matrix.workers.map((worker) => [worker.slug, worker]));
  const definitions = catalog.worker_types.map((entry) => {
    const definition = readJson(path.join(root, entry.definition_path));
    return { definition, evidence: evidence.get(entry.slug) };
  });
  requireComplete(definitions);
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, render(definitions));
  console.log(`wrote Worker lifecycle templates: ${output}`);
}

function render(definitions) {
  const lines = [
    "# Worker Lifecycle Templates",
    "",
    "Generated from `config/worker-types/*/definition.json` and the current Worker evidence matrix.",
    "",
    "## Common Release Sequence",
    "",
    "1. Build or locate the exact runtime image, then run its declared version probe.",
    "2. Fetch create options, select only Definition-declared model and credential references, and preflight the draft.",
    "3. In the browser, verify required fields, validation errors, and type-switch reset behavior without exposing a secret value.",
    "4. With a named disposable target, create one Worker, prove the declared PTY or ACP connection, then run one harmless prompt.",
    "5. Verify termination, Runner cleanup, browser state, and console/network errors before changing that Worker to supported.",
    "",
    "## Per-Worker Templates",
    "",
  ];
  for (const item of definitions.sort((left, right) =>
    left.definition.slug.localeCompare(right.definition.slug),
  )) {
    lines.push(...renderWorker(item.definition, item.evidence), "");
  }
  return lines.join("\n");
}

function renderWorker(definition, evidence) {
  const details = runtimeDetails[definition.slug];
  const lifecycle = definition.interaction_modes.includes("acp")
    ? "ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup."
    : "PTY attach, harmless prompt, terminal output, terminate, and cleanup.";
  return [
    `### ${definition.slug}`,
    "",
    `- Build: ${details.build}`,
    `- Definition: executable \`${definition.executable}\`; adapter \`${definition.adapter_id}\`; modes \`${definition.interaction_modes.join("/")}\`.`,
    `- Form: ${modelRequirement(definition)} ${credentialRequirement(definition)} ${documentRequirement(definition)} ${toolModelRequirement(definition)}`,
    `- Runner: ${details.runner}`,
    `- Current state: ${runtimeState(evidence)}.`,
    `- Lifecycle proof: ${lifecycle}`,
    `- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.`,
  ];
}

function modelRequirement(definition) {
  const requirement = definition.model_requirement;
  if (!requirement.required) return "No primary model resource is required.";
  return `Primary model adapters: \`${requirement.protocol_adapters.join(", ")}\`.`;
}

function credentialRequirement(definition) {
  if (definition.credential_bindings.length === 0) return "No credential binding.";
  const values = definition.credential_bindings.map((binding) =>
    `${binding.source.kind}:${binding.source.ref}->${binding.target.name}`,
  );
  return `Credential bindings: \`${values.join(", ")}\`.`;
}

function documentRequirement(definition) {
  if (definition.config_documents.length === 0) return "No Definition-owned config document.";
  const values = definition.config_documents.map((document) =>
    `${document.id}:${document.format}->${document.target_path}`,
  );
  return `Named config document bindings required: \`${values.join(", ")}\`.`;
}

function toolModelRequirement(definition) {
  if (!definition.tool_model_requirements?.length) return "No tool model.";
  const values = definition.tool_model_requirements.map((requirement) =>
    `${requirement.id}:${requirement.provider_keys.join(",")}/${requirement.modality}/${requirement.capability}`,
  );
  return `Tool model: \`${values.join(", ")}\`.`;
}

function runtimeState(evidence) {
  if (!evidence) return "missing evidence row";
  if (evidence.runtime.evidence_state === "blocked") {
    return `blocked: ${evidence.runtime.create_option_blocking_reason}`;
  }
  return "runtime evidence only; no successful lifecycle is recorded";
}

function requireComplete(definitions) {
  for (const item of definitions) {
    if (!item.definition.slug || !item.evidence || !runtimeDetails[item.definition.slug]) {
      throw new Error("Worker lifecycle template input is incomplete");
    }
  }
}

function outputPath(argumentsList) {
  if (argumentsList.length === 0) return defaultOutput;
  if (argumentsList.length === 2 && argumentsList[0] === "--output") {
    return path.resolve(argumentsList[1]);
  }
  throw new Error(
    "usage: node render-worker-lifecycle-templates.mjs [--output <path>]",
  );
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}
