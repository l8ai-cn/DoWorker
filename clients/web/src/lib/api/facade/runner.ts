import { getRunnerService } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";
import { create as protoCreate, toBinary, fromBinary } from "@bufbuild/protobuf";
import {
  AuthorizeRunnerRequestSchema,
  AuthorizeRunnerResponseSchema,
  GetRunnerAuthStatusRequestSchema,
  RunnerAuthStatusSchema,
} from "@proto/runner_api/v1/runner_pb";
import {
  listRunners as listRunnersConnect,
  listAvailableRunners as listAvailableRunnersConnect,
  getRunner as getRunnerConnect,
  updateRunner as updateRunnerConnect,
  deleteRunner as deleteRunnerConnect,
  createRunnerToken as createRunnerTokenConnect,
  listRunnerLogs as listRunnerLogsConnect,
  querySandboxes as querySandboxesConnect,
  requestLogUpload as requestLogUploadConnect,
  upgradeRunner as upgradeRunnerConnect,
} from "../connect/runnerConnect";
import { listPods as listPodsConnect } from "../connect/podConnect";

export type { RunnerData, GRPCRegistrationToken, RunnerPodData, SandboxStatus, RelayConnectionInfo, RunnerLogData } from "@/lib/viewModels/runner";

export interface RunnerAuthStatus {
  status: string;
  runner_id?: number;
  node_id?: string;
  organization_name?: string;
}

function orgSlug(): string {
  return readCurrentOrg()?.slug ?? "";
}

export const runnerApi = {
  list: async (status?: string) => {
    const { items, total, limit, offset, latest_runner_version } = await listRunnersConnect(orgSlug(), { status });
    return { runners: items, total, limit, offset, latest_runner_version };
  },
  get: async (id: number) => {
    return await getRunnerConnect(orgSlug(), id);
  },
  fetchRunners: async (status?: string) => {
    const { items, total, limit, offset, latest_runner_version } = await listRunnersConnect(orgSlug(), { status });
    return { runners: items, total, limit, offset, latest_runner_version };
  },
  fetchRunner: async (id: number) => {
    return await getRunnerConnect(orgSlug(), id);
  },
  fetchAvailableRunners: async () => {
    const { items, total } = await listAvailableRunnersConnect(orgSlug());
    return { runners: items, total };
  },
  update: async (id: number, data: Record<string, unknown>) => {
    return await updateRunnerConnect(orgSlug(), id, data);
  },
  delete: async (id: number) => {
    await deleteRunnerConnect(orgSlug(), id);
  },
  createToken: async (data: { cluster_id: number; name?: string; labels?: string[]; max_uses?: number; expires_in_days?: number }) => {
    return await createRunnerTokenConnect(orgSlug(), data);
  },
  listLogs: async (id: number) => {
    const { items, total, limit, offset } = await listRunnerLogsConnect(orgSlug(), id);
    return { logs: items, total, limit, offset };
  },
  // List pods filtered by runner_id via proto.pod.v1.PodService.ListPods.
  // (The dedicated REST endpoint was retired; ListPods has a runner_id
  // filter we delegate to here.)
  listPods: async (id: number, filters?: { status?: string; limit?: number; offset?: number }) => {
    const { items, total } = await listPodsConnect(orgSlug(), {
      runner_id: id, status: filters?.status, limit: filters?.limit, offset: filters?.offset,
    });
    return { items, total };
  },
  querySandboxes: async (id: number, podKeys: string[]) => {
    return await querySandboxesConnect(orgSlug(), id, podKeys);
  },
  requestLogUpload: async (id: number) => {
    await requestLogUploadConnect(orgSlug(), id);
  },
  upgrade: async (id: number, data?: Record<string, unknown>) => {
    const targetVersion = (data?.target_version as string) ?? "";
    return await upgradeRunnerConnect(orgSlug(), id, targetVersion);
  },
};

// get_auth_status / authorize_runner aren't part of proto.runner_api.v1's
// renderer-facing surface today, but the wire format IS proto — the bridge
// just speaks GetRunnerAuthStatus / AuthorizeRunner directly so renderer
// encodes/decodes via @bufbuild/protobuf rather than the legacy JSON shim.
export const runnerAuthApi = {
  getAuthStatus: async (authKey: string): Promise<RunnerAuthStatus> => {
    const req = protoCreate(GetRunnerAuthStatusRequestSchema, { authKey });
    const respBytes = await getRunnerService().get_auth_status(
      toBinary(GetRunnerAuthStatusRequestSchema, req),
    );
    const resp = fromBinary(RunnerAuthStatusSchema, new Uint8Array(respBytes));
    return {
      status: resp.status,
      runner_id: resp.runnerId !== undefined ? Number(resp.runnerId) : undefined,
      node_id: resp.nodeId,
      organization_name: resp.orgSlug,
    };
  },
  authorize: async (data: { auth_key: string; node_id?: string; cluster_id: number }) => {
    const req = protoCreate(AuthorizeRunnerRequestSchema, {
      authKey: data.auth_key,
      nodeId: data.node_id ?? "",
      clusterId: BigInt(data.cluster_id),
      // orgSlug filled in by the service layer from session if empty.
      orgSlug: orgSlug(),
    });
    const respBytes = await getRunnerService().authorize_runner(
      toBinary(AuthorizeRunnerRequestSchema, req),
    );
    fromBinary(AuthorizeRunnerResponseSchema, new Uint8Array(respBytes));
  },
};
