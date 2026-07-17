import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  CreateRegistrationCommandRequestSchema,
  CreateRegistrationCommandResponseSchema,
  ListExecutionClustersRequestSchema,
  ListExecutionClustersResponseSchema,
} from "@proto/execution_cluster/v1/execution_cluster_pb";
import { getExecutionClusterService } from "@/lib/wasm-core";
import {
  fromExecutionCluster,
  type ExecutionCluster,
  type RegistrationCommand,
} from "../facade/executionCluster";

export async function listExecutionClusters(
  orgSlug: string,
): Promise<ExecutionCluster[]> {
  const request = create(ListExecutionClustersRequestSchema, { orgSlug });
  const bytes = await getExecutionClusterService().listExecutionClustersConnect(
    toBinary(ListExecutionClustersRequestSchema, request),
  );
  return fromBinary(
    ListExecutionClustersResponseSchema,
    new Uint8Array(bytes),
  ).items.map(fromExecutionCluster);
}

export async function createRegistrationCommand(
  orgSlug: string,
  clusterId: number,
  nodeName = "",
): Promise<RegistrationCommand> {
  if (!Number.isSafeInteger(clusterId) || clusterId <= 0) {
    throw new Error("invalid execution cluster id");
  }
  const request = create(CreateRegistrationCommandRequestSchema, {
    orgSlug,
    clusterId: BigInt(clusterId),
    nodeName,
  });
  const bytes =
    await getExecutionClusterService().createRegistrationCommandConnect(
      toBinary(CreateRegistrationCommandRequestSchema, request),
    );
  const response = fromBinary(
    CreateRegistrationCommandResponseSchema,
    new Uint8Array(bytes),
  );
  return { command: response.command, expiresAt: response.expiresAt };
}
