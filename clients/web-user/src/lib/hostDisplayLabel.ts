import type { Host } from "@/hooks/useHosts";
import { sandboxOptionLabel } from "@/lib/agent-cloud/server-info";

const INTERNAL_RUNNER_HOST_RE = /^admin-workspace-/i;

function isInternalRunnerSlug(slug: string): boolean {
  return INTERNAL_RUNNER_HOST_RE.test(slug);
}

export function isInternalWorkspaceRunnerHost(host: Host): boolean {
  const slug = host.host_id.replace(/^host_/, "");
  return isInternalRunnerSlug(host.name) || isInternalRunnerSlug(slug);
}

function isInternalRunnerHost(host: Host): boolean {
  return isInternalWorkspaceRunnerHost(host);
}

export function unresolvedHostLabel(hostId: string): string {
  const slug = hostId.replace(/^host_/, "");
  if (isInternalRunnerSlug(slug)) return "Workspace";
  return hostId;
}

export function hostDisplayLabel(
  host: Host,
  options?: { thisMachineHostId?: string | null },
): string {
  if (options?.thisMachineHostId && host.host_id === options.thisMachineHostId) {
    return "This machine";
  }
  if (host.sandbox_provider) {
    return sandboxOptionLabel(host.sandbox_provider);
  }
  if (isInternalRunnerHost(host)) {
    return "Workspace";
  }
  return host.name;
}
