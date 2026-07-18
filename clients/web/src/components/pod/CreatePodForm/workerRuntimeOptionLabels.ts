import type { WorkerRuntimeSelectOption } from "./WorkerRuntimeSelectField";

type RuntimeOptionKind =
  | "workerType"
  | "runtimeImage"
  | "computeTarget"
  | "deploymentMode"
  | "resourceProfile";

type Translate = (key: string) => string;

const workerTypeLabels: Record<string, string> = {
  "codex-cli": "workerCreate.runtime.options.codex",
  "claude-code": "workerCreate.runtime.options.claude",
  "gemini-cli": "workerCreate.runtime.options.gemini",
  "minimax-cli": "workerCreate.runtime.options.minimax",
  openclaw: "workerCreate.runtime.options.openclaw",
  "do-agent": "workerCreate.runtime.options.doAgent",
  "seedance-expert": "workerCreate.runtime.options.seedance",
  "pattern-designer": "workerCreate.runtime.options.patternDesigner",
  aider: "workerCreate.runtime.options.aider",
  opencode: "workerCreate.runtime.options.openCode",
  cursor: "workerCreate.runtime.options.cursor",
};

const reasonKeys: Record<string, string> = {
  "No runtime image is available for this worker type":
    "workerCreate.runtime.options.noRuntimeImage",
  "Dedicated managed Kubernetes provisioning is not configured":
    "workerCreate.runtime.options.dedicatedUnavailable",
  "Compute target is disabled": "workerCreate.runtime.options.computeTargetDisabled",
  "No enabled compute target supports this deployment mode":
    "workerCreate.runtime.options.noTargetForMode",
  "Selected compute target is unavailable":
    "workerCreate.runtime.options.selectedTargetUnavailable",
  "Selected compute target does not support this deployment mode":
    "workerCreate.runtime.options.targetDoesNotSupportMode",
  "Resource profile is disabled": "workerCreate.runtime.options.resourceDisabled",
  "Dedicated provisioning is disabled":
    "workerCreate.runtime.options.dedicatedUnavailable",
};

export function localizeWorkerRuntimeOption(
  kind: RuntimeOptionKind,
  value: string,
  name: string,
  selectable: boolean,
  blockingReason: string,
  t: Translate,
  lookupValue = value,
): WorkerRuntimeSelectOption {
  return {
    value,
    label: localizeLabel(kind, lookupValue, name, t),
    selectable,
    blockingReason: localizeReason(blockingReason, t),
  };
}

function localizeLabel(
  kind: RuntimeOptionKind,
  value: string,
  name: string,
  t: Translate,
): string {
  const key = labelKey(kind, value);
  if (key) return t(key);
  return name.replace(/\s+\(local development\)$/i, "");
}

function labelKey(kind: RuntimeOptionKind, value: string): string | undefined {
  if (kind === "workerType") return workerTypeLabels[value];
  if (kind === "computeTarget") {
    if (value === "organization-runner-pool") {
      return "workerCreate.runtime.options.organizationRunnerPool";
    }
    if (value === "managed-kubernetes") {
      return "workerCreate.runtime.options.managedKubernetes";
    }
  }
  if (kind === "deploymentMode") {
    if (value === "pooled") return "workerCreate.runtime.options.pooled";
    if (value === "dedicated") return "workerCreate.runtime.options.dedicated";
  }
  if (kind === "resourceProfile") {
    if (value === "standard" || value === "1") {
      return "workerCreate.runtime.options.standard";
    }
    if (value === "large" || value === "2") {
      return "workerCreate.runtime.options.large";
    }
  }
  return undefined;
}

function localizeReason(reason: string, t: Translate): string {
  const key = reasonKeys[reason];
  return key ? t(key) : reason;
}
