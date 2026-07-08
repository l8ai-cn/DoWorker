import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import type { Host } from "@/hooks/useHosts";
import { isInternalWorkspaceRunnerHost } from "./hostDisplayLabel";

function harnessReported(host: Host, key: string): boolean | string | undefined {
  return host.configured_harnesses?.[key];
}

export function hostSupportsAgent(
  host: Host | undefined,
  agent: AvailableAgent | null | undefined,
): boolean {
  if (!host || !agent || host.status !== "online") return false;
  const configured = host.configured_harnesses;
  if (!configured || Object.keys(configured).length === 0) return true;
  const byId = harnessReported(host, agent.id);
  if (byId === true) return true;
  if (byId === false) return false;
  if (agent.harness) {
    const byHarness = harnessReported(host, agent.harness);
    if (byHarness === true) return true;
    if (byHarness === false) return false;
  }
  return false;
}

export function pickOnlineHostForAgent(
  hosts: Host[],
  agent: AvailableAgent | null | undefined,
): Host | undefined {
  const online = hosts.filter((h) => h.status === "online");
  if (online.length === 0) return undefined;
  if (!agent) return online[0];
  const supporting = online.filter((h) => hostSupportsAgent(h, agent));
  return supporting.length > 0 ? supporting[0] : undefined;
}

export function hostsShareInternalWorkspaceGroup(a: Host, b: Host): boolean {
  return isInternalWorkspaceRunnerHost(a) && isInternalWorkspaceRunnerHost(b);
}

/** One picker row for all admin-workspace-* runners (dev ships several). */
export function collapseInternalWorkspaceHostsForPicker(
  hosts: Host[],
  agent?: AvailableAgent | null,
): Host[] {
  const internal = hosts.filter(isInternalWorkspaceRunnerHost);
  if (internal.length <= 1) return hosts;
  const external = hosts.filter((h) => !isInternalWorkspaceRunnerHost(h));
  const onlineInternal = internal.filter((h) => h.status === "online");
  const pool = onlineInternal.length > 0 ? onlineInternal : internal;
  const representative = pickOnlineHostForAgent(pool, agent) ?? pool[0];
  return [...external, representative];
}

export function isHostActiveInPicker(
  host: Host,
  selectedHostId: string | null,
  allHosts: Host[],
): boolean {
  if (!selectedHostId) return false;
  if (host.host_id === selectedHostId) return true;
  const selected = allHosts.find((h) => h.host_id === selectedHostId);
  if (!selected) return false;
  return hostsShareInternalWorkspaceGroup(host, selected);
}
