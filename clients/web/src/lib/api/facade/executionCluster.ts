import type { ExecutionCluster as ProtoExecutionCluster } from "@proto/execution_cluster/v1/execution_cluster_pb";

export interface ExecutionCluster {
  id: number;
  slug: string;
  name: string;
  kind: string;
  status: string;
  runnerCount: number;
  onlineRunnerCount: number;
  availableRunnerCount: number;
  tunnelStatus: string;
  tunnelLastSeenAt?: string;
  tunnelLastError?: string;
}

export interface RegistrationCommand {
  command: string;
  expiresAt: string;
}

function safeId(value: bigint): number {
  const id = Number(value);
  if (!Number.isSafeInteger(id)) throw new Error("unsafe execution cluster id");
  return id;
}

export function fromExecutionCluster(
  value: ProtoExecutionCluster,
): ExecutionCluster {
  return {
    id: safeId(value.id),
    slug: value.slug,
    name: value.name,
    kind: value.kind,
    status: value.status,
    runnerCount: value.runnerCount,
    onlineRunnerCount: value.onlineRunnerCount,
    availableRunnerCount: value.availableRunnerCount,
    tunnelStatus: value.tunnelStatus,
    tunnelLastSeenAt: value.tunnelLastSeenAt,
    tunnelLastError: value.tunnelLastError,
  };
}
