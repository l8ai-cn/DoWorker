import { lightConnect } from "./api-fetch";

export interface LightExecutionCluster {
  id: string;
  slug: string;
  name: string;
  kind: string;
  status: string;
  runnerCount?: number;
  onlineRunnerCount?: number;
  availableRunnerCount?: number;
  tunnelStatus?: string;
  tunnelLastSeenAt?: string;
  tunnelLastError?: string;
}

interface ConnectExecutionCluster {
  id: number | string;
  slug: string;
  name: string;
  kind: string;
  status: string;
  runnerCount?: number;
  onlineRunnerCount?: number;
  availableRunnerCount?: number;
  tunnelStatus?: string;
  tunnelLastSeenAt?: string;
  tunnelLastError?: string;
}

export async function lightListExecutionClusters(
  orgSlug: string,
): Promise<LightExecutionCluster[]> {
  const response = await lightConnect<
    { orgSlug: string },
    { items?: ConnectExecutionCluster[] }
  >(
    "proto.execution_cluster.v1.ExecutionClusterService",
    "ListExecutionClusters",
    { orgSlug },
    { authenticated: true },
  );
  return (response.items ?? []).map((cluster) => ({
    ...cluster,
    id: exactClusterID(cluster.id),
  }));
}

function exactClusterID(value: number | string): string {
  if (typeof value === "number") {
    if (!Number.isSafeInteger(value) || value <= 0) {
      throw new Error("unsafe execution cluster id");
    }
    return String(value);
  }
  if (!/^[1-9]\d*$/.test(value)) {
    throw new Error("invalid execution cluster id");
  }
  return value;
}
