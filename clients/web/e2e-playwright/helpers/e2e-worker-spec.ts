import type { ConnectClient } from "./connect-client";
import { E2E_ECHO_AGENT_SLUG } from "./e2e-echo-runner";
import { TEST_ORG_SLUG } from "./env";
import { createE2EPodAlias, registerE2ECreatedPod } from "./pod-cleanup";
type InteractionMode = "pty" | "acp";
type AutomationLevel = "interactive" | "auto_edit" | "autonomous";
interface WorkerTypeOption {
  slug: string;
  schemaVersion: number;
  configSchemaJson: string;
  selectable: boolean;
  blockingReason?: string;
  requiresModelResource: boolean;
  supportedInteractionModes: string[];
}

interface RuntimeImageOption {
  id: bigint; selectable: boolean; workerTypeSlugs: string[];
}
interface ComputeTargetOption {
  id: bigint;
  kind: string;
  supportsPooled: boolean;
  selectable: boolean;
}

interface ResourceProfileOption {
  id: bigint; slug: string; selectable: boolean;
}
interface WorkerCreateOptions {
  revision: string;
  workerTypes: WorkerTypeOption[];
  runtimeImages: RuntimeImageOption[];
  computeTargets: ComputeTargetOption[];
  deploymentModes: Array<{ value: string; selectable: boolean }>;
  resourceProfiles: ResourceProfileOption[];
}

export interface E2EWorkerSpecOptions {
  mode?: InteractionMode;
  automationLevel?: AutomationLevel;
  scenario?: string;
  prompt?: string;
  alias?: string;
  repositoryId?: bigint;
  branch?: string;
  envBundleIds?: bigint[];
}

export interface E2ECreatePodOptions extends E2EWorkerSpecOptions {
  ticketSlug?: string;
  cols?: number;
  rows?: number;
}

export interface E2EWorkerSpecDraft {
  workerTypeSlug: string;
  runtimeImageId: bigint;
  placementPolicy: string;
  computeTargetId: bigint;
  deploymentMode: string;
  resourceProfileId: bigint;
  typeSchemaVersion: number;
  typeConfigValuesJson: string;
  interactionMode: string;
  automationLevel: string;
  repositoryId?: bigint;
  branch: string;
  envBundleIds: bigint[];
  instructions: string;
  initialTask: string;
  terminationPolicy: string;
  idleTimeoutMinutes: number;
  alias: string;
  optionsRevision: string;
}

export async function buildE2EEchoWorkerSpec(
  client: ConnectClient,
  options: E2EWorkerSpecOptions = {},
): Promise<E2EWorkerSpecDraft> {
  const branch = options.branch?.trim() ?? "";
  if (options.repositoryId !== undefined && !branch) {
    throw new Error("repository-backed WorkerSpec requires an explicit branch");
  }
  const catalog = await client.pod.listWorkerCreateOptions({
    orgSlug: TEST_ORG_SLUG,
    workerTypeSlug: E2E_ECHO_AGENT_SLUG,
  }) as WorkerCreateOptions;
  const workerType = catalog.workerTypes.find(
    (item) => item.slug === E2E_ECHO_AGENT_SLUG && item.selectable,
  );
  if (!workerType) {
    const candidate = catalog.workerTypes.find(
      (item) => item.slug === E2E_ECHO_AGENT_SLUG,
    );
    throw new Error(
      candidate?.blockingReason ||
        "ListWorkerCreateOptions did not expose selectable e2e-echo",
    );
  }
  if (workerType.requiresModelResource) {
    throw new Error("e2e-echo unexpectedly requires a model resource");
  }
  const mode = options.mode ?? "pty";
  if (!workerType.supportedInteractionModes.includes(mode)) {
    throw new Error(`e2e-echo does not support ${mode} interaction`);
  }
  const runtime = requiredOption(
    catalog.runtimeImages,
    (item) => item.selectable &&
      item.workerTypeSlugs.includes(E2E_ECHO_AGENT_SLUG),
    "selectable e2e-echo runtime image",
  );
  const target = requiredOption(
    catalog.computeTargets,
    (item) => item.selectable && item.kind === "runner-pool" &&
      item.supportsPooled,
    "selectable pooled runner target",
  );
  requiredOption(
    catalog.deploymentModes,
    (item) => item.selectable && item.value === "pooled",
    "selectable pooled deployment mode",
  );
  const profile = catalog.resourceProfiles.find(
    (item) => item.selectable && item.slug === "standard",
  ) ?? requiredOption(
    catalog.resourceProfiles,
    (item) => item.selectable,
    "selectable resource profile",
  );
  return {
    workerTypeSlug: E2E_ECHO_AGENT_SLUG,
    runtimeImageId: runtime.id,
    placementPolicy: "automatic",
    computeTargetId: target.id,
    deploymentMode: "pooled",
    resourceProfileId: profile.id,
    typeSchemaVersion: workerType.schemaVersion,
    typeConfigValuesJson: JSON.stringify(
      workerTypeConfig(workerType.configSchemaJson, options.scenario),
    ),
    interactionMode: mode,
    automationLevel: options.automationLevel ?? "interactive",
    repositoryId: options.repositoryId,
    branch,
    envBundleIds: options.envBundleIds ?? [],
    instructions: "",
    initialTask: options.prompt ?? "",
    terminationPolicy: "manual",
    idleTimeoutMinutes: 0,
    alias: createE2EPodAlias(options.alias),
    optionsRevision: catalog.revision,
  };
}

export async function createE2EEchoPod(
  client: ConnectClient,
  options: E2ECreatePodOptions = {},
) {
  const workerSpec = await buildE2EEchoWorkerSpec(client, options);
  const created = await client.pod.createPod({
    orgSlug: TEST_ORG_SLUG,
    ticketSlug: options.ticketSlug,
    cols: options.cols ?? 80,
    rows: options.rows ?? 24,
    workerSpec,
  });
  const podKey = (created as { pod?: { podKey?: string } }).pod?.podKey;
  if (!podKey) throw new Error("CreatePod returned no pod key for E2E cleanup registration");
  registerE2ECreatedPod(podKey, workerSpec.alias);
  return created;
}

function workerTypeConfig(raw: string, scenario?: string) {
  const schema = JSON.parse(raw || "{}") as {
    fields?: Record<string, { default?: unknown; required?: boolean }>;
  };
  const fields = schema.fields ?? {};
  const values: Record<string, unknown> = {};
  for (const [name, field] of Object.entries(fields)) {
    if ("default" in field) values[name] = field.default;
    else if (field.required) {
      throw new Error(`e2e-echo config ${name} has no default`);
    }
  }
  if (scenario && scenario !== "echo") {
    if (!fields.scenario) {
      throw new Error("e2e-echo Worker definition does not declare scenario");
    }
    values.scenario = scenario;
  }
  return values;
}

function requiredOption<T>(
  items: T[],
  predicate: (item: T) => boolean,
  label: string,
): T {
  const match = items.find(predicate);
  if (!match) throw new Error(`ListWorkerCreateOptions missing ${label}`);
  return match;
}
