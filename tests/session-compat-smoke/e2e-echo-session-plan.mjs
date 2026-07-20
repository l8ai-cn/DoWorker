const API = process.env.SESSION_COMPAT_API_URL || "http://localhost:10015";
const ORG = "dev-org";
const WORKER_TYPE = "e2e-echo";

export async function buildE2EEchoSessionBody(token, input = {}) {
  const plan = await buildWorkerPlan(token);
  const hostId = input.host_id ?? await selectHost(token);
  const ptyOnly = input.pty_only === true;
  return {
    ...input,
    agent_id: WORKER_TYPE,
    host_id: hostId,
    workspace: input.workspace ?? "/workspace",
    ...plan,
    automation_level: ptyOnly ? "interactive" : "autonomous",
  };
}

export async function createE2EEchoSession(token, input = {}) {
  const body = await buildE2EEchoSessionBody(token, input);
  const res = await fetch(`${API}/v1/sessions`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(body),
  });
  const data = await jsonBody(res);
  if (!res.ok) {
    throw new Error(`create session ${res.status}: ${JSON.stringify(data)}`);
  }
  if (typeof data.id !== "string" || data.id.length === 0) {
    throw new Error(`create session returned no id: ${JSON.stringify(data)}`);
  }
  return data;
}

async function buildWorkerPlan(token) {
  const initial = await listWorkerCreateOptions(token, { workerTypeSlug: WORKER_TYPE });
  const revision = requiredRevision(initial.revision);
  const worker = select(initial.workerTypes, (item) => item.slug === WORKER_TYPE, "selectable e2e-echo worker");
  if (!worker.supportedInteractionModes?.includes("acp")) {
    throw new Error("e2e-echo does not support ACP sessions");
  }
  const compute = select(
    initial.computeTargets,
    (item) => item.kind === "runner-pool" && item.supportsPooled,
    "selectable pooled runner target",
  );
  const deploymentOptions = await listWorkerCreateOptions(token, {
    workerTypeSlug: WORKER_TYPE,
    computeTargetId: optionID(compute.id, "compute target id"),
  });
  assertRevision(revision, deploymentOptions.revision);
  const deployment = select(
    deploymentOptions.deploymentModes,
    (item) => item.value === "pooled",
    "selectable pooled deployment mode",
  );
  const resolved = await listWorkerCreateOptions(token, {
    workerTypeSlug: WORKER_TYPE,
    computeTargetId: optionID(compute.id, "compute target id"),
    deploymentMode: deployment.value,
  });
  assertRevision(revision, resolved.revision);
  const runtime = select(
    resolved.runtimeImages,
    (item) => item.workerTypeSlugs?.includes(WORKER_TYPE),
    "selectable e2e-echo runtime image",
  );
  const profile = resolved.resourceProfiles.find(
    (item) => item.selectable && item.slug === "standard",
  ) ?? select(resolved.resourceProfiles, () => true, "selectable resource profile");
  return {
    worker_spec: {
      options_revision: revision,
      runtime_image_id: optionID(runtime.id, "runtime image id"),
      placement_policy: "automatic",
      compute_target_id: optionID(compute.id, "compute target id"),
      deployment_mode: deployment.value,
      resource_profile_id: optionID(profile.id, "resource profile id"),
    },
  };
}

async function listWorkerCreateOptions(token, filter) {
  const body = {
    orgSlug: ORG,
    workerTypeSlug: filter.workerTypeSlug,
  };
  if (filter.computeTargetId !== undefined) {
    body.computeTargetId = String(filter.computeTargetId);
  }
  if (filter.deploymentMode !== undefined) {
    body.deploymentMode = filter.deploymentMode;
  }
  const res = await fetch(`${API}/proto.pod.v1.PodService/ListWorkerCreateOptions`, {
    method: "POST",
    headers: {
      ...authHeaders(token),
      "Connect-Protocol-Version": "1",
    },
    body: JSON.stringify(body),
  });
  const data = await jsonBody(res);
  if (!res.ok) {
    throw new Error(`ListWorkerCreateOptions ${res.status}: ${JSON.stringify(data)}`);
  }
  return data;
}

async function selectHost(token) {
  const res = await fetch(`${API}/v1/hosts`, { headers: authHeaders(token) });
  const body = await jsonBody(res);
  if (!res.ok) throw new Error(`list hosts ${res.status}: ${JSON.stringify(body)}`);
  const hosts = Array.isArray(body.hosts) ? body.hosts : [];
  const candidates = hosts.filter((host) =>
    host.status === "online" && host.configured_harnesses?.[WORKER_TYPE] === true
  );
  candidates.sort((a, b) => scoreHost(b) - scoreHost(a));
  const selected = candidates[0];
  if (!selected?.host_id) {
    throw new Error("no online e2e-echo host is available");
  }
  return selected.host_id;
}

function scoreHost(host) {
  if (host.host_id === "host_dev-runner-2") return 2;
  if (host.host_id === "host_dev-runner") return 1;
  return 0;
}

function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": ORG,
    "Content-Type": "application/json",
  };
}

async function jsonBody(res) {
  const text = await res.text();
  if (!text) return {};
  try {
    return JSON.parse(text);
  } catch {
    return { raw: text };
  }
}

function select(items, predicate, label) {
  const item = (items ?? []).find((candidate) => candidate.selectable && predicate(candidate));
  if (!item) throw new Error(`ListWorkerCreateOptions missing ${label}`);
  return item;
}

function optionID(value, label) {
  const result = typeof value === "bigint" ? Number(value) : Number(value);
  if (!Number.isSafeInteger(result) || result <= 0) {
    throw new Error(`invalid ${label}`);
  }
  return result;
}

function requiredRevision(value) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error("Worker create options revision is missing");
  }
  return value;
}

function assertRevision(expected, actual) {
  if (expected !== actual) {
    throw new Error("Worker create options changed while preparing session");
  }
}
